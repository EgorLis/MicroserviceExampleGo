package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/domain/events"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type PaymentsRepo struct {
	pool *pgxpool.Pool
}

func NewPaymentsRepo(dsn string) (*PaymentsRepo, error) {
	pool, err := newPostgresPool(dsn)
	if err != nil {
		return nil, err
	}
	return &PaymentsRepo{pool: pool}, nil
}

func (r *PaymentsRepo) Ping() error {
	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping error: %w", err)
	}
	return nil
}

func (r *PaymentsRepo) Close() {
	r.pool.Close()
	log.Println("Postgres closed...")
}

// ToDo: Пофиксить первые миграции, когда нет схемы...
func (r *PaymentsRepo) RunMigrations() error {
	ctx := context.Background()

	// получаем текущую версию схемы
	var currentVersion int
	err := r.pool.QueryRow(ctx, "SELECT version FROM provider.meta ORDER BY version DESC LIMIT 1").
		Scan(&currentVersion)
	if err != nil {
		// если таблицы meta нет -> создаём
		if strings.Contains(err.Error(), "does not exist") {
			_, err = r.pool.Exec(ctx, `
                CREATE TABLE IF NOT EXISTS provider.meta (
                    version int not null,
                    applied_at timestamptz not null default now()
                );
            `)
			if err != nil {
				return err
			}
			currentVersion = 0
		} else {
			return err
		}
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// сортировка по имени файла
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for idx, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}

		migrationVersion := idx + 1 // допустим, просто нумеруем 001_xxx.sql -> 1
		if migrationVersion <= currentVersion {
			continue // уже применяли
		}

		sqlBytes, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return err
		}

		fmt.Println(">> executing", e.Name())
		if _, err := r.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", e.Name(), err)
		}

		// записываем новую версию
		_, err = r.pool.Exec(ctx,
			`INSERT INTO provider.meta(version) VALUES($1)`,
			migrationVersion,
		)
		if err != nil {
			return err
		}

		currentVersion = migrationVersion
	}

	return nil
}

func newPostgresPool(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN failed: %w", err)
	}
	// Настроим таймауты
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("parse DSN failed: %w", err)
	}

	// Проверим соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("parse DSN failed: %w", err)
	}

	return pool, nil
}

func (r *PaymentsRepo) InsertProcessedEvent(ctx context.Context, payment events.PaymentProcessed) error {
	res, err := r.pool.Exec(ctx, `
    	INSERT INTO provider.processed_events (payment_id, status, psp_reference)
    	VALUES ($1,$2,$3)
    	ON CONFLICT (payment_id) DO NOTHING`,
		payment.PaymentID, payment.Status, payment.PSPRef)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		// значит, запись с таким payment_id уже есть
		log.Printf("postgres: duplicate processed event, payment_id: %s", payment.PaymentID)
	}

	return nil
}

func (r *PaymentsRepo) Statistic(ctx context.Context) (events.Statistic, error) {
	stats := events.Statistic{}
	err := r.pool.QueryRow(ctx, `
	SELECT
    	COUNT(*) 										AS processed,
    	COUNT(*) FILTER (WHERE status = 'AUTHORIZED')   AS authorized,
    	COUNT(*) FILTER (WHERE status = 'DECLINED')     AS declined,
	FROM provider.processed_events;
	`).Scan(
		&stats.Processed,
		&stats.Authorized,
		&stats.Declined)

	if err != nil {
		return events.Statistic{}, nil
	}

	return stats, nil
}
