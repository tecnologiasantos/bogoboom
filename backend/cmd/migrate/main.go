// Command migrate applies backend/migrations/*.sql in order and records them in schema_migrations.
// Usage: go run ./cmd/migrate [-dir migrations]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/joho/godotenv"

	"bogoboom-backend/internal/config"
)

func main() {
	dir := flag.String("dir", "migrations", "folder with the .sql files")
	flag.Parse()
	_ = godotenv.Load()

	ctx := context.Background()
	pool, err := config.NewDatabasePool(ctx, config.LoadDatabaseConfig())
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		log.Fatalf("schema_migrations: %v", err)
	}
	files, err := filepath.Glob(filepath.Join(*dir, "*.sql"))
	if err != nil || len(files) == 0 {
		log.Fatalf("no migrations found in %s", *dir)
	}
	sort.Strings(files)
	for _, f := range files {
		name := filepath.Base(f)
		var done bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, name).Scan(&done); err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		if done {
			continue
		}
		sql, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("%s failed: %v", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			log.Fatalf("%s: %v", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("%s: %v", name, err)
		}
		fmt.Println("applied", name)
	}
	fmt.Println("migrations up to date")
}
