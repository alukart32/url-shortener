package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/alukart32/shortener-url/internal/shortener/storage/shortenedurl"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// shortURLSaver represents the shortURL saver for the postgres repository.
type shortURLSaver struct {
	pool *pgxpool.Pool
}

// ShortURLSaver returns a new shortURLSaver.
func ShortURLSaver(pool *pgxpool.Pool) *shortURLSaver {
	return &shortURLSaver{
		pool: pool,
	}
}

// Save inserts a new shortURL. A successful call returns err == nil.
func (s *shortURLSaver) Save(ctx context.Context, data models.ShortenedURL) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.RepeatableRead,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()

	const insertShortURL = `INSERT INTO
	shorturls(slug, user_id, original, short, corr_id)
	VALUES ($1, $2, $3, $4, $5)`

	_, err = tx.Exec(ctx, insertShortURL,
		data.Slug,
		data.UserID,
		data.Raw,
		data.Value,
		data.CorrID,
	)

	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.SQLState()) &&
			pgErr.SQLState() == pgerrcode.UniqueViolation {
			return shortenedurl.ErrUniqueViolation
		}
	}
	return err
}

// Batch performs bulk shortURL insertion. It uses the PostgreSQL copy protocol to perform bulk data insertion.
// A successful call returns err == nil.
func (s *shortURLSaver) Batch(ctx context.Context, records []models.ShortenedURL) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.RepeatableRead,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()

	// Prepare CopyForm. It uses the PostgreSQL copy protocol to perform bulk data insertion.
	_, err = tx.CopyFrom(ctx,
		pgx.Identifier{"shorturls"},
		[]string{"slug", "user_id", "original", "short", "corr_id"},
		pgx.CopyFromSlice(len(records), func(i int) ([]any, error) {
			return []interface{}{
				records[i].Slug,
				records[i].UserID,
				records[i].Raw,
				records[i].Value,
				records[i].CorrID,
			}, nil
		}),
	)

	var pgErr *pgconn.PgError
	if err != nil && errors.As(err, &pgErr) {
		if pgerrcode.IsIntegrityConstraintViolation(pgErr.SQLState()) &&
			pgErr.SQLState() == pgerrcode.UniqueViolation {
			err = fmt.Errorf("%v", pgErr.Message)
		}
	}
	return err
}
