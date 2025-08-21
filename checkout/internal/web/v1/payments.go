package v1

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/config"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/events"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/idempotency"
	"github.com/EgorLis/MicroserviceExampleGo/checkout/internal/domain/payment"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
)

type PaymentsHandler struct {
	Repo      payment.Repository
	IdemStore idempotency.Store
	Publisher events.Publisher
	Cfg       *config.Config
}

func (ph *PaymentsHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10000*time.Millisecond)
	defer cancel()

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		writeError(w, http.StatusBadRequest, "idempotency key required")
		return
	}

	if err := validateIdempotencyKey(idemKey); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req paymentCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	defer r.Body.Close()

	if errs := validatePayment(req); len(errs) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	bodyHash, err := canonicalHash(req)
	if err != nil {
		log.Printf("idempotency error: %v", err)
		writeError(w, http.StatusInternalServerError, "")
		return
	}

	// 1) пробуем создать «резервацию» (SETNX)
	created, err := ph.IdemStore.Reserve(ctx, req.MerchantID, idemKey, bodyHash, idempotency.TTL)
	if err != nil {
		// Проверка на timeout / отмену контекста
		if isTimeout(err) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
			return
		}
		// Прочие ошибки
		log.Printf("idempotency error: %v", err)
		writeError(w, http.StatusInternalServerError, "")
		return
	}

	if !created {
		// ключ уже был
		val, err := ph.IdemStore.Load(ctx, req.MerchantID, idemKey)
		if err != nil {
			// Проверка на timeout / отмену контекста
			if isTimeout(err) {
				writeError(w, http.StatusGatewayTimeout, "request timed out")
				return
			}
			// Прочие ошибки
			log.Printf("idempotency error: %v", err)
			writeError(w, http.StatusInternalServerError, "")
			return
		}
		if val.BodyHash != bodyHash {
			writeError(w, http.StatusUnprocessableEntity, "idempotency key reused with different payload")
			return
		}

		switch val.State {
		case idempotency.StateInProgress:
			existPayment, err := ph.Repo.GetPaymentByUniqKeys(ctx, req.MerchantID, req.OrderID)
			var resp PaymentCreateResponse
			if err != nil {
				// Проверка на timeout / отмену контекста
				if isTimeout(err) {
					writeError(w, http.StatusGatewayTimeout, "request timed out")
					return
				}
				// Прочие ошибки
				log.Println(err)
				if errors.Is(err, pgx.ErrNoRows) {
					resp.Status = string(payment.StatusProcessing)
					writeJSON(w, http.StatusAccepted, resp)
					return
				}
				log.Printf("db error: %v", err)
				writeError(w, http.StatusInternalServerError, "")
				return
			}
			resp.PaymentID = existPayment.ID
			resp.Status = string(existPayment.Status)
			code := http.StatusCreated
			err = ph.IdemStore.Finalize(ctx, req.MerchantID, idemKey, bodyHash, code, existPayment.ID,
				map[string]any{"payment_id": resp.PaymentID, "status": resp.Status}, idempotency.TTL)
			if err != nil {
				// Проверка на timeout / отмену контекста
				if isTimeout(err) {
					writeError(w, http.StatusGatewayTimeout, "request timed out")
					return
				}
				// Прочие ошибки
				writeError(w, http.StatusInternalServerError, "")
				return
			}

			log.Printf("idempotency: value finalized")

			writeJSON(w, code, resp)
			return
		case idempotency.StateDone:
			writeJSON(w, val.HTTPCode, val.Response)
			return
		case idempotency.StateError:
			writeError(w, http.StatusInternalServerError, "previous attempt failed")
			return
		}

	}

	log.Printf("idempotency: value reserved")

	payID := createPaymentID()

	amount, _ := decimal.NewFromString(req.Amount)
	pay := payment.Payment{
		ID:          payID,
		MerchantID:  req.MerchantID,
		OrderID:     req.OrderID,
		Amount:      amount,
		Currency:    req.Currency,
		MethodToken: req.MethodToken,
		Status:      payment.StatusPending,
	}

	// db logic
	if err := ph.Repo.InsertPayment(ctx, pay); err != nil {
		// 1. Проверка на timeout / отмену контекста
		if isTimeout(err) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
			return
		}

		// 2. Проверка на уникальность
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "payment already exists")
			return
		}

		// 3. Всё остальное → 500
		log.Printf("db error: %v", err)
		writeError(w, http.StatusInternalServerError, "")
		return
	}

	log.Println("db: row added")

	resp := PaymentCreateResponse{Status: string(pay.Status), PaymentID: payID}
	code := http.StatusCreated
	// 3) записать финализацию в Redis и обновить TTL
	err = ph.IdemStore.Finalize(ctx, req.MerchantID, idemKey, bodyHash, code, payID,
		map[string]any{"payment_id": resp.PaymentID, "status": resp.Status}, idempotency.TTL)
	if err != nil {
		// Проверка на timeout / отмену контекста
		if isTimeout(err) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
			return
		}
		// Прочие ошибки
		log.Printf("idempotency error: %v", err)
		writeError(w, http.StatusInternalServerError, "")
		return
	}

	log.Printf("idempotency: value finalized")

	event, err := events.NewPaymentCreatedEvent(pay)
	event.Headers["x-idempotency-key"] = idemKey
	event.Headers["x-trace-id"] = uuid.NewString()

	if err != nil {
		log.Printf("kafka: build event failed payment_id=%s err=%v", pay.ID, err)
	}

	if err = ph.Publisher.Publish(ctx, event); err != nil {
		log.Printf("kafka: publish failed payment_id=%s err=%v", pay.ID, err)
	} else {
		log.Printf("kafka: published payment_id=%s", pay.ID)
	}

	writeJSON(w, code, resp)
}

func (ph *PaymentsHandler) Get(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	paymentID := parts[3]

	if !validatePayID(paymentID) {
		writeError(w, http.StatusBadRequest, "wrong id")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), ph.Cfg.HTTP.PaymentTimeout)
	defer cancel()

	payment, err := ph.Repo.GetPaymentByID(ctx, paymentID)
	if err != nil {
		// Проверка на timeout / отмену контекста
		if isTimeout(err) {
			writeError(w, http.StatusGatewayTimeout, "request timed out")
			return
		}
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "")
		return
	}

	resp := ToResponse(payment)

	writeJSON(w, http.StatusOK, resp)
}
