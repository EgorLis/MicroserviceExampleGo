package postgres

import "github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"

// db -> domain
func ToDomain(row PaymentRow) payment.Payment {
	return payment.Payment{
		ID: row.ID, MerchantID: row.MerchantID,
		OrderID: row.OrderID, Amount: row.Amount,
		Currency: row.Currency, Status: payment.PaymentStatus(row.Status),
		PSPRef: row.PSPRef, MethodToken: row.MethodToken,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// domain -> row
func ToRow(p payment.Payment) PaymentRow {
	return PaymentRow{
		ID: p.ID, MerchantID: p.MerchantID,
		OrderID: p.OrderID, Amount: p.Amount,
		Currency: p.Currency, Status: string(p.Status),
		PSPRef: p.PSPRef, MethodToken: p.MethodToken,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
