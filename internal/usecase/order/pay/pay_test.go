package pay_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	. "github.com/tiptop-co/backend/internal/usecase/order/pay"
)

func TestUsecase_Pay(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name         string
		paymentID    int64
		amount       int64
		prepare      func(payment *MockpaymentRepo)
		expectations func(t assert.TestingT, err error)
	}{
		{},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockPaymentRepo := NewMockpaymentRepo(ctrl)
			
			if tc.prepare != nil {
				tc.prepare(mockPaymentRepo)
			}
			
			instance := New(mockPaymentRepo)

			err := instance.Pay(context.Background(), tc.paymentID, tc.amount)
			
			tc.expectations(t, err)
		})
	}
}
