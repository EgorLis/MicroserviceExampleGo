package v1

type PaymentCreateResponse struct {
	PaymentID string `json:"payment_id,omitempty"`
	Status    string `json:"status"`
}

type PaymentResponse struct {
	ID         string  `json:"payment_id"`
	MerchantID string  `json:"merchant_id"`
	OrderID    string  `json:"order_id"`
	Amount     string  `json:"amount"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	PSPRef     *string `json:"psp_reference"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type healthResponse struct {
	Status string `json:"status"`
}
type versionResponse struct {
	Version string `json:"version"`
}
