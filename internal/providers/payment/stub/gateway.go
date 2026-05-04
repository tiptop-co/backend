package payment_stub

import (
	"context"

	"github.com/tiptop-co/backend/internal/model/transaction"
)

type PaymentGateway interface {
	Charge(ctx context.Context, transactionID string, amount int) (transaction.Status, error)
}

type StubGateway struct{}

func New() *StubGateway {
	return &StubGateway{}
}

func (g *StubGateway) Charge(_ context.Context, _ string, _ int) (transaction.Status, error) {
	return transaction.StatusSuccess, nil
}
