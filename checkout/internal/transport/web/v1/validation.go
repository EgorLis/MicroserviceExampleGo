package v1

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var currencyRe = regexp.MustCompile(`^[A-Z]{3}$`)

var iso4217 = map[string]struct{}{
	"USD": {},
	"EUR": {},
	"RUB": {},
}

func validatePayment(req paymentCreateRequest) []string {
	var errs []string

	if !validateString(req.OrderID) {
		errs = append(errs, "invalid order_id")
	}
	if !validateString(req.MethodToken) {
		errs = append(errs, "invalid method_token")
	}
	if !validateString(req.MerchantID) {
		errs = append(errs, "invalid merchant_id")
	}
	if !validateDecimal(req.Amount) {
		errs = append(errs, "invalid amount")
	}
	if !validateCurrency(req.Currency) {
		errs = append(errs, "invalid currency")
	}

	return errs
}

func validateIdempotencyKey(key string) error {
	key = strings.TrimSpace(key)
	l := utf8.RuneCountInString(key)
	if l == 0 || l > 64 {
		return errors.New("idempotency key must have atleast 1 symbol and less 65")
	}
	return nil
}

func validateCurrency(code string) bool {
	if !currencyRe.MatchString(code) {
		return false
	}
	_, ok := iso4217[code]
	return ok
}

func validateString(s string) bool {
	s = strings.TrimSpace(s)
	l := utf8.RuneCountInString(s)
	return l > 0 && l <= 128
}

func validateDecimal(dec string) bool {
	d, err := decimal.NewFromString(dec)
	if err != nil {
		return false
	}
	// число должно быть > 0
	if d.Cmp(decimal.Zero) <= 0 {
		return false
	}
	// не больше 2 знаков после запятой
	if d.Exponent() < -2 {
		return false
	}
	return true
}

func validatePayID(s string) bool {
	if !strings.HasPrefix(s, "pay_") {
		return false
	}
	idPart := strings.TrimPrefix(s, "pay_")

	_, err := uuid.Parse(idPart)
	return err == nil
}
