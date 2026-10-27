// Package repository implements the domain ports on PostgreSQL with pgx.
//
// Conventions: IDs are exchanged as text (`id::text` out, `$n::text::uuid` in), money as float8 and
// dates as YYYY-MM-DD text, so the domain stays free of driver types.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the connection pool the repositories share.
type DB = pgxpool.Pool

// scanner is satisfied by pgx.Row and pgx.Rows.
type scanner interface{ Scan(dest ...any) error }

// querier is satisfied by *pgxpool.Pool and pgx.Tx.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// inTx runs fn in a transaction: commit on nil, rollback on error.
func inTx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after Commit
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// isUniqueViolation reports a duplicate-key error (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
