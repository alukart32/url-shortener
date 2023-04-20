package postgres

import (
	"context"
	"errors"

	"github.com/alukart32/shortener-url/internal/shortener/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// shortURLProvider represents the shortURL provider for the postgres repository.
type shortURLProvider struct {
	pool *pgxpool.Pool
}

// ShortURLProvider returns a new shortURLProvider.
func ShortURLProvider(pool *pgxpool.Pool) *shortURLProvider {
	return &shortURLProvider{
		pool: pool,
	}
}

// GetBySlug finds a shortURL by slug. A successful call returns err == nil.
func (p *shortURLProvider) GetBySlug(ctx context.Context, slug string) (models.ShortenedURL, error) {
	const getBySlug = `SELECT * FROM shorturls WHERE slug = $1`

	var r models.ShortenedURL
	row := p.pool.QueryRow(ctx, getBySlug, slug)
	err := row.Scan(
		&r.Slug,
		&r.UserID,
		&r.Raw,
		&r.Value,
		&r.CorrID,
		&r.IsDeleted,
	)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}

	return r, err
}

// GetByURL returns a shortURL by original URL.
func (p *shortURLProvider) GetByURL(ctx context.Context, url string) (models.ShortenedURL, error) {
	const getByURL = `SELECT * FROM shorturls WHERE original = $1`

	var r models.ShortenedURL
	row := p.pool.QueryRow(ctx, getByURL, url)
	err := row.Scan(
		&r.Slug,
		&r.UserID,
		&r.Raw,
		&r.Value,
		&r.CorrID,
		&r.IsDeleted,
	)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}

	return r, err
}

// Collect finds shortURLs by using userID. A successful call returns err == nil.
func (p *shortURLProvider) CollectByUser(ctx context.Context, userID string) ([]models.ShortenedURL, error) {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.RepeatableRead,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.TODO())
		} else {
			tx.Commit(context.TODO())
		}
	}()

	const listByUserID = `SELECT * FROM shorturls WHERE user_id = $1`

	rows, err := tx.Query(ctx, listByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]models.ShortenedURL, 0)
	for rows.Next() {
		var r models.ShortenedURL
		if err = rows.Scan(
			&r.Slug,
			&r.UserID,
			&r.Raw,
			&r.Value,
			&r.CorrID,
			&r.IsDeleted,
		); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return records, err
}
