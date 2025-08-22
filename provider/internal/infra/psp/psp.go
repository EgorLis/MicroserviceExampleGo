package psp

import (
	"math/rand/v2"

	"github.com/EgorLis/MicroserviceExampleGo/provider/internal/config"
	"github.com/google/uuid"
)

type PSPStatus string

const (
	Authorized PSPStatus = "AUTHORIZED"
	Declined   PSPStatus = "DECLINED"
)

type Simulator struct {
	cfg config.PSP
}

func New(cfg *config.PSP) *Simulator {
	return &Simulator{
		cfg: *cfg,
	}
}

func (s *Simulator) DecidePayment() (status string, pspRef *string) {
	if isAuthorized(s.cfg.Chance) {
		ref := s.cfg.Prefix + uuid.NewString() // генерируйте как угодно
		return string(Authorized), &ref
	}
	return string(Declined), nil
}

func isAuthorized(chance float64) bool {
	return rand.Float64() < chance // true ~80% случаев
}
