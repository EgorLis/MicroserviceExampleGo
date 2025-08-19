package db

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(dsn string) (*PostgresDB, error) {
	pool, err := newPostgresPool(dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Ping() error {
	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping error: %v", err)
	}
	return nil
}

func (db *PostgresDB) Close() {
	db.pool.Close()
	log.Println("Postgres closed...")
}

func (db *PostgresDB) RunMigrations() error {
	ctx := context.Background()

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// сортировка по имени файла
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		sqlBytes, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}

		fmt.Println(">> executing", e.Name())
		if _, err := db.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", e.Name(), err)
		}
	}

	return nil
}

func newPostgresPool(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN failed: %v", err)
	}
	// Настроим таймауты
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("parse DSN failed: %v", err)
	}

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("parse DSN failed: %v", err)
	}

	return pool, nil
}
