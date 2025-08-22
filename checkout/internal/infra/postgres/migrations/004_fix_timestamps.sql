-- Приводим created_at / updated_at к UTC по умолчанию
ALTER TABLE checkout.payments
    ALTER COLUMN created_at SET DEFAULT (now() at time zone 'utc'),
    ALTER COLUMN updated_at SET DEFAULT (now() at time zone 'utc');

-- Обновляем существующие данные
UPDATE checkout.payments
SET created_at = created_at AT TIME ZONE 'utc',
    updated_at = updated_at AT TIME ZONE 'utc';