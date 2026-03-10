package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var (
		cmd    = flag.String("cmd", "up", "up|down|seed|status")
		steps  = flag.Int("steps", 1, "steps for down")
		mDir   = flag.String("dir", "migrations", "migration directory")
		sDir   = flag.String("seed", "seed", "seed directory")
		dbURL  = flag.String("db", os.Getenv("DATABASE_URL"), "database url")
	)
	flag.Parse()

	if *dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, *dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := ensureSchemaTable(ctx, pool); err != nil {
		log.Fatal(err)
	}

	switch *cmd {
	case "up":
		if err := migrateUp(ctx, pool, *mDir); err != nil {
			log.Fatal(err)
		}
	case "down":
		if err := migrateDown(ctx, pool, *mDir, *steps); err != nil {
			log.Fatal(err)
		}
	case "seed":
		if err := runSeed(ctx, pool, *sDir); err != nil {
			log.Fatal(err)
		}
	case "status":
		ver, err := currentVersion(ctx, pool)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("current_version=%d\n", ver)
	default:
		log.Fatalf("unknown cmd: %s", *cmd)
	}
}

func ensureSchemaTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`)
	return err
}

func currentVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var v int
	err := pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&v)
	return v, err
}

func migrateUp(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil {
		return err
	}
	migs, err := parseMigrations(files, ".up.sql")
	if err != nil {
		return err
	}
	cur, err := currentVersion(ctx, pool)
	if err != nil {
		return err
	}
	for _, m := range migs {
		if m.Version <= cur {
			continue
		}
		if err := applyMigration(ctx, pool, m.Path, m.Version); err != nil {
			return err
		}
		fmt.Printf("applied %d\n", m.Version)
	}
	return nil
}

func migrateDown(ctx context.Context, pool *pgxpool.Pool, dir string, steps int) error {
	if steps < 1 {
		return errors.New("steps must be >= 1")
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.down.sql"))
	if err != nil {
		return err
	}
	migs, err := parseMigrations(files, ".down.sql")
	if err != nil {
		return err
	}
	cur, err := currentVersion(ctx, pool)
	if err != nil {
		return err
	}
	if cur == 0 {
		return nil
	}
	var toApply []migration
	for i := len(migs) - 1; i >= 0; i-- {
		if migs[i].Version <= cur {
			toApply = append(toApply, migs[i])
		}
	}
	if len(toApply) < steps {
		steps = len(toApply)
	}
	for i := 0; i < steps; i++ {
		m := toApply[i]
		if err := applyDown(ctx, pool, m.Path, m.Version); err != nil {
			return err
		}
		fmt.Printf("rolled back %d\n", m.Version)
	}
	return nil
}

func runSeed(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, f := range files {
		if err := execSQLFile(ctx, pool, f); err != nil {
			return err
		}
		fmt.Printf("seeded %s\n", filepath.Base(f))
	}
	return nil
}

type migration struct {
	Version int
	Path    string
}

func parseMigrations(files []string, suffix string) ([]migration, error) {
	var res []migration
	for _, f := range files {
		base := filepath.Base(f)
		parts := strings.Split(base, "_")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid migration name: %s", base)
		}
		vstr := parts[0]
		v, err := strconv.Atoi(vstr)
		if err != nil {
			return nil, fmt.Errorf("invalid version in %s", base)
		}
		if !strings.HasSuffix(base, suffix) {
			continue
		}
		res = append(res, migration{Version: v, Path: f})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Version < res[j].Version })
	return res, nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, path string, version int) error {
	sql, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if _, err = tx.Exec(ctx, string(sql)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func applyDown(ctx context.Context, pool *pgxpool.Pool, path string, version int) error {
	sql, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	if _, err = tx.Exec(ctx, string(sql)); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM schema_migrations WHERE version=$1`, version); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func execSQLFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	sql, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = pool.Exec(ctx, string(sql))
	return err
}
