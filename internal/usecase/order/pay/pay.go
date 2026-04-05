package pay

import (
	"context"
	"fmt"

	"github.com/tiptop-co/backend/internal/model"
)

type Usecase struct {
	payment paymentRepo
}

func New(p paymentRepo) *Usecase {
	return &Usecase{
		payment: p,
	}
}

func (u *Usecase) Pay(ctx context.Context, paymentID int64, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	err := u.payment.UpdateStatus(ctx, paymentID, model.PaymentStatusPaid)
	if err != nil {
		return fmt.Errorf("order pay usecase: failed to update status: %w", err)
	}

	return nil
}
