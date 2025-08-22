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
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/shared/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PickBatch: статус NEW/FAILED, время пришло, отметим IN_PROGRESS и вернём.
const pickSQL = `
WITH cte AS (
  SELECT id
  FROM checkout.outbox_events
  WHERE status IN ('NEW','FAILED') AND next_attempt_at <= now()
  ORDER BY id
  FOR UPDATE SKIP LOCKED
  LIMIT $1
)
UPDATE checkout.outbox_events o
SET status='IN_PROGRESS', updated_at=now()
FROM cte
WHERE o.id = cte.id
RETURNING o.id, o.event_type, o.key, o.payload, o.headers;
`

const resetSQL = `
UPDATE checkout.outbox_events o
SET status='FAILED',
    attempt = attempt + 1,
	next_attempt_at = now() + interval '30 sec',
    updated_at=now()
WHERE o.updated_at < now() - interval '5 min' AND
	o.status = 'IN_PROGRESS'
`

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

func (r *PaymentsRepo) InsertPayment(ctx context.Context, payment payment.Payment, env event.Envelope) error {
	eventRow, err := EnvelopeToRow(env)
	if err != nil {
		return fmt.Errorf("invalid event, can't parse to row %w", err)
	}
	payRow := PaymentToRow(payment)

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // безопасно: если уже коммитнули — no-op

	_, err = tx.Exec(ctx,
		`INSERT INTO checkout.payments (payment_id, merchant_id, order_id, amount, currency, method_token, psp_reference)
  		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		payRow.ID, payRow.MerchantID, payRow.OrderID, payRow.Amount, payRow.Currency, payRow.MethodToken, payRow.PSPRef)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO checkout.outbox_events (aggregate_type, aggregate_id, event_type, key, payload, headers)
  		 VALUES ($1,$2,$3,$4,$5,$6)`,
		eventRow.AggregateType, eventRow.AggregateID, eventRow.EventType,
		eventRow.Key, eventRow.Payload, eventRow.Headers)

	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
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

	return PaymentRowToDomain(row), err
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

	return PaymentRowToDomain(row), err
}

func (r *PaymentsRepo) PickBatch(ctx context.Context, count int) (map[int64]event.Envelope, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // безопасно: если уже коммитнули — no-op

	rows, err := tx.Query(ctx, pickSQL, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	envs := make(map[int64]event.Envelope, count)

	for rows.Next() {
		var outRow OutboxEventRow
		err := rows.Scan(&outRow.ID, &outRow.EventType, &outRow.Key, &outRow.Payload, &outRow.Headers)
		if err != nil {
			return nil, fmt.Errorf("cant parse row to outboxEventRow, err:%w", err)
		}
		env, err := OutboxRowToEnvelope(outRow)
		if err != nil {
			return nil, fmt.Errorf("cant parse outboxEventRow to envelope, err:%w", err)
		}
		envs[outRow.ID] = env
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error with rows: %w", rows.Err())
	}

	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}

	return envs, nil
}

func (r *PaymentsRepo) MarkSent(ctx context.Context, ids []int64) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE checkout.outbox_events
        SET status='SENT', updated_at=now()
        WHERE id = ANY($1)
    `, ids)
	return err
}

func (r *PaymentsRepo) MarkFailed(ctx context.Context, ids []int64) error {
	_, err := r.pool.Exec(ctx, `
        UPDATE checkout.outbox_events
        SET status='FAILED',
            attempt = attempt + 1,
            next_attempt_at = now() + interval '30 sec',
            updated_at = now()
        WHERE id = ANY($1)
    `, ids)
	return err
}

func (r *PaymentsRepo) ResetEvents(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, resetSQL); err != nil {
		return err
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
