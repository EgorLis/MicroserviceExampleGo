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

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
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
		return fmt.Errorf("ping error: %v", err)
	}
	return nil
}

func (r *PaymentsRepo) Close() {
	r.pool.Close()
	log.Println("Postgres closed...")
}

func (r *PaymentsRepo) RunMigrations() error {
	ctx := context.Background()

	// получаем текущую версию схемы
	var currentVersion int
	err := r.pool.QueryRow(ctx, "SELECT version FROM checkout.meta ORDER BY version DESC LIMIT 1").
		Scan(&currentVersion)
	if err != nil {
		// если таблицы meta нет -> создаём
		if strings.Contains(err.Error(), "does not exist") {
			_, err = r.pool.Exec(ctx, `
                CREATE TABLE IF NOT EXISTS checkout.meta (
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
			`INSERT INTO checkout.meta(version) VALUES($1)`,
			migrationVersion,
		)
		if err != nil {
			return err
		}

		currentVersion = migrationVersion
	}

	return nil
}

func (r *PaymentsRepo) InsertPayment(ctx context.Context, payment payment.Payment) error {
	row := ToRow(payment)

	_, err := r.pool.Exec(ctx,
		`INSERT INTO checkout.payments (payment_id, merchant_id, order_id, amount, currency, method_token, psp_reference)
  		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		row.ID, row.MerchantID, row.OrderID, row.Amount, row.Currency, row.MethodToken, row.PSPRef)

	return err
}

func (r *PaymentsRepo) GetPaymentByID(ctx context.Context, id string) (payment.Payment, error) {
	var row PaymentRow

	err := r.pool.QueryRow(ctx,
		`SELECT payment_id, merchant_id, order_id, amount, currency, status, psp_reference, created_at, updated_at
         FROM checkout.payments
         WHERE payment_id = $1`, id,
	).Scan(
		&row.ID,
		&row.MerchantID,
		&row.OrderID,
		&row.Amount,
		&row.Currency,
		&row.Status,
		&row.PSPRef,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	return ToDomain(row), err
}

func (r *PaymentsRepo) GetPaymentByUniqKeys(ctx context.Context, merchantID, orderID string) (payment.Payment, error) {
	var row PaymentRow

	err := r.pool.QueryRow(ctx,
		`SELECT payment_id, merchant_id, order_id, amount, currency, status, psp_reference, created_at, updated_at
         FROM checkout.payments
         WHERE merchant_id = $1 AND order_id = $2`, merchantID, orderID,
	).Scan(
		&row.ID,
		&row.MerchantID,
		&row.OrderID,
		&row.Amount,
		&row.Currency,
		&row.Status,
		&row.PSPRef,
		&row.CreatedAt,
		&row.UpdatedAt,
	)

	return ToDomain(row), err
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
