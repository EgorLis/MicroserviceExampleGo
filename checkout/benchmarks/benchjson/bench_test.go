package benchjson

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

// ---------- данные теста ----------
type Status string

const (
	StatusPending Status = "PENDING"
)

type PaymentCreated struct {
	EventType    string `json:"event_type"`
	EventVersion int    `json:"event_version"`
	PaymentID    string `json:"payment_id"`
	MerchantID   string `json:"merchant_id"`
	OrderID      string `json:"order_id"`
	Amount       string `json:"amount"` // в проде часто decimal->string
	Currency     string `json:"currency"`
	Status       Status `json:"status"`
	OccurredAt   string `json:"occurred_at"`
}

var (
	paymentID  = "pay_bc342cbc-8da0-4016-80e7-3967557df853"
	merchantID = "m_129"
	orderID    = "o_456"
	amountStr  = "100.00"
	currency   = "USD"
	nowISO     = time.Now().UTC().Format(time.RFC3339Nano)

	// заранее подготовленные объекты (для «чистой» сериализации)
	eventStruct = PaymentCreated{
		EventType:    "PaymentCreated",
		EventVersion: 1,
		PaymentID:    paymentID,
		MerchantID:   merchantID,
		OrderID:      orderID,
		Amount:       amountStr,
		Currency:     currency,
		Status:       StatusPending,
		OccurredAt:   nowISO,
	}
	eventMap = map[string]any{
		"event_type":    "PaymentCreated",
		"event_version": 1,
		"payment_id":    paymentID,
		"merchant_id":   merchantID,
		"order_id":      orderID,
		"amount":        amountStr,
		"currency":      currency,
		"status":        string(StatusPending),
		"occurred_at":   nowISO,
	}
)

// ---------- 1) сериализация только (готовые объекты) ----------

func BenchmarkJSON_Marshal_Struct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(eventStruct)
	}
}

func BenchmarkJSON_Marshal_Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(eventMap)
	}
}

// ---------- 2) построение + сериализация (ближе к реальности) ----------

func BenchmarkBuildAndMarshal_Struct(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ev := PaymentCreated{
			EventType:    "PaymentCreated",
			EventVersion: 1,
			PaymentID:    paymentID,
			MerchantID:   merchantID,
			OrderID:      orderID,
			Amount:       amountStr,
			Currency:     currency,
			Status:       StatusPending,
			OccurredAt:   nowISO,
		}
		_, _ = json.Marshal(ev)
	}
}

func BenchmarkBuildAndMarshal_Map(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ev := map[string]any{
			"event_type":    "PaymentCreated",
			"event_version": 1,
			"payment_id":    paymentID,
			"merchant_id":   merchantID,
			"order_id":      orderID,
			"amount":        amountStr,
			"currency":      currency,
			"status":        string(StatusPending),
			"occurred_at":   nowISO,
		}
		_, _ = json.Marshal(ev)
	}
}

// ---------- 3) encoder + переиспользуемый bytes.Buffer ----------

func BenchmarkEncoder_Struct_ReusedBuffer(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.Encode(eventStruct) // пишет с \n в конце
	}
}

func BenchmarkEncoder_Map_ReusedBuffer(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = enc.Encode(eventMap)
	}
}

// ---------- 4) вариант с «сырым amount» (если заранее есть []byte) ----------

// полезно, если ты заранее формируешь amount как []byte и хочешь уменьшить аллокации
type PaymentCreatedRawAmount struct {
	EventType    string    `json:"event_type"`
	EventVersion int       `json:"event_version"`
	PaymentID    string    `json:"payment_id"`
	MerchantID   string    `json:"merchant_id"`
	OrderID      string    `json:"order_id"`
	Amount       RawString `json:"amount"` // сериализуется как строка без копирований
	Currency     string    `json:"currency"`
	Status       Status    `json:"status"`
	OccurredAt   string    `json:"occurred_at"`
}

// RawString — легковесный тип, который реализует json.Marshaler,
// чтобы писать уже готовые байты строки как JSON-строку.
type RawString []byte

func (r RawString) MarshalJSON() ([]byte, error) {
	// оборачиваем в кавычки и экранируем, если нужно.
	// для простоты предположим, что amount безопасен (только [0-9.]),
	// тогда можно написать минимально:
	out := make([]byte, 0, len(r)+2)
	out = append(out, '"')
	out = append(out, r...)
	out = append(out, '"')
	return out, nil
}

var eventStructRaw = PaymentCreatedRawAmount{
	EventType:    "PaymentCreated",
	EventVersion: 1,
	PaymentID:    paymentID,
	MerchantID:   merchantID,
	OrderID:      orderID,
	Amount:       RawString([]byte(amountStr)),
	Currency:     currency,
	Status:       StatusPending,
	OccurredAt:   nowISO,
}

func BenchmarkJSON_Marshal_Struct_RawAmount(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(eventStructRaw)
	}
}
