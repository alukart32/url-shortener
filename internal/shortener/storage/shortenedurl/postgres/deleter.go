package postgres

import (
	"context"
	"runtime"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// shortURLDeleter represents the shortURL batcher for the postgres repository.
type shortURLDeleter struct {
	pool *pgxpool.Pool
	mtx  sync.Mutex
}

// ShortURLDeleter returns a new shortURLDeleter for shortURLs.
func ShortURLDeleter(pool *pgxpool.Pool) *shortURLDeleter {
	return &shortURLDeleter{
		pool: pool,
	}
}

// Delete sets shortURL deleted status.
//
// The update takes place using a set of batches, that are compiled by worker pool.
func (r *shortURLDeleter) Delete(userID string, slugs []string) error {

	// Set the pool size.
	wcount := runtime.NumCPU()
	if len(slugs)%5 == 0 {
		wcount = 5
	} else if len(slugs) < wcount {
		wcount = len(slugs)
	}

	slugCh := make(chan string, wcount)
	go func() {
		defer close(slugCh)
		for _, v := range slugs {
			slugCh <- v
		}
	}()

	const markAsDeleted = `UPDATE shorturls SET deleted = true WHERE slug = $1 and user_id = $2`

	batch := pgx.Batch{}
	var mtx sync.Mutex

	var wg sync.WaitGroup
	wg.Add(wcount)
	for i := 0; i < wcount; i++ {
		go func() {
			defer wg.Done()

			for slug := range slugCh {
				mtx.Lock()
				batch.Queue(markAsDeleted, slug, userID)
				mtx.Unlock()
			}
		}()
	}

	wg.Wait()

	txCtx := context.Background()
	tx, err := r.pool.BeginTx(txCtx, pgx.TxOptions{
		IsoLevel:       pgx.RepeatableRead,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(txCtx)
		} else {
			tx.Commit(txCtx)
		}
	}()

	var batchResult pgx.BatchResults
	r.mtx.Lock()
	batchResult = tx.SendBatch(txCtx, &batch)
	r.mtx.Unlock()
	return batchResult.Close()
}
