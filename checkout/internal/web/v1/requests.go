package v1

type paymentCreateRequest struct {
	MerchantID  string `json:"merchant_id"`
	OrderID     string `json:"order_id"`
	Amount      string `json:"amount"`
	Currency    string `json:"currency"`
	MethodToken string `json:"method_token"`
}
