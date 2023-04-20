package postgres

import (
	"context"
	"fmt"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// statProvider represents the shortenedURLs statistics provider for the postgres repository.
type statProvider struct {
	pool *pgxpool.Pool
}

// StatProvider returns a new statProvider.
func StatProvider(pool *pgxpool.Pool) *statProvider {
	return &statProvider{
		pool: pool,
	}
}

// Stat collects statistics about shortened URLs.
func (sp *statProvider) Stat(ctx context.Context) (models.Stat, error) {
	var (
		tx  pgx.Tx
		err error
	)

	tx, err = sp.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	},
	)
	if err != nil {
		return models.Stat{}, fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		err = sp.finishTx(ctx, tx, err)
	}()

	const getStat = `
SELECT
	SUM(users.cnt) AS urls,
	COUNT(*) AS users
FROM
(
	SELECT
		COUNT(*) AS cnt
	FROM shorturls
	GROUP BY user_id
) AS users
`

	row := tx.QueryRow(ctx, getStat)

	var stat models.Stat
	if err := row.Scan(&stat.URLs, &stat.Users); err != nil {
		return models.Stat{}, fmt.Errorf("failed to scan pgx.Row: %v", err)
	}

	return stat, nil
}

// finishTx rollbacks transaction if error is provided. If err is nil transaction is committed.
func (sp *statProvider) finishTx(ctx context.Context, tx pgx.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("%v, failed to rollback tx: %v", err, rollbackErr)
		}

		return err
	} else {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return fmt.Errorf("failed to commit tx: %v", commitErr)
		}

		return nil
	}
}
