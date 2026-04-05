package payment

// НЕ ИСПОЛЬЗОВАТЬ
// Все содержимое данного файла является примером!

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/tiptop-co/backend/internal/model"
	paymentModel "github.com/tiptop-co/backend/internal/model/payment"
)

func (r *Repository) UpdateStatus(ctx context.Context, paymentID int64, status string) (paymentModel.Payment, error) {
	query := `
		update payment
		set status = $1
		where payment_id = $2;
	`

	var payment paymentModel.Payment

	err := r.db.GetContext(ctx, &payment, query, status, paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return paymentModel.Payment{}, model.ErrNotFound
	}
	if err != nil {
		return paymentModel.Payment{}, fmt.Errorf("item repository.GetByID: %w", err)
	}

	return payment, nil
}
