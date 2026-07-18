// Package postgres wires sqlc-generated queries to a pgx pool and adds
// transaction support. It implements the Store/Querier seams consumed by usecases.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/alamjad/marketplace/internal/repository/postgres/sqlc"
)

// Repository embeds the generated Queries (so it satisfies sqlc.Querier) and
// keeps the pool for running transactions.
type Repository struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

// New builds a Repository from a pgx pool.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{Queries: sqlc.New(pool), pool: pool}
}

// WithTx runs fn within a database transaction, committing on success.
func (r *Repository) WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit
	if err := fn(r.Queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
