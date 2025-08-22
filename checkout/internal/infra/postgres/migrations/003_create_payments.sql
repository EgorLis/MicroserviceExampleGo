CREATE TYPE checkout.payment_status AS ENUM ('PENDING', 'SUCCEEDED', 'FAILED');

CREATE TABLE IF NOT EXISTS checkout.payments (
    payment_id    text PRIMARY KEY,
    merchant_id   text NOT NULL,
    order_id      text NOT NULL,
    amount        numeric(20,2) NOT NULL CHECK (amount > 0),
    currency      text NOT NULL CHECK (char_length(currency) = 3),
    method_token  text NOT NULL,
    status        checkout.payment_status NOT NULL DEFAULT 'PENDING',
    psp_reference text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- уникальность платежа на стороне мерчанта (пока без идемпотентности ключом)
CREATE UNIQUE INDEX IF NOT EXISTS ux_checkout_payments_merchant_order
ON checkout.payments(merchant_id, order_id);