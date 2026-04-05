//go:generate mockgen -source ${GOFILE} -destination mocks_test.go -package ${GOPACKAGE}_test
package pay

import (
	"context"

	"github.com/tiptop-co/backend/internal/model"
)

type paymentRepo interface {
	UpdateStatus(ctx context.Context, paymentID int64, status model.PaymentStatus) error
}
