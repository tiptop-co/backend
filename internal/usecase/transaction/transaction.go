package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/tiptop-co/backend/internal/model"
	ordermodel "github.com/tiptop-co/backend/internal/model/order"
	tablemodel "github.com/tiptop-co/backend/internal/model/table"
	tx_model "github.com/tiptop-co/backend/internal/model/transaction"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[TRANSACTION USECASE]"

type TransactionService struct {
	orders OrderRepository
	txRepo TransactionRepository
	tables TableRepo
	pay    PaymentGateway
}

func NewTransactionService(orders OrderRepository, txRepo TransactionRepository, tables TableRepo, pay PaymentGateway) *TransactionService {
	return &TransactionService{orders: orders, txRepo: txRepo, tables: tables, pay: pay}
}

func (s *TransactionService) Create(ctx context.Context, orderID string, itemIDs []string, tipsAmount int) (_ *tx_model.Transaction, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create", err) }()

	if orderID == "" {
		return nil, model.ErrValidation
	}
	if len(itemIDs) == 0 {
		return nil, ErrEmptyItemList
	}
	if tipsAmount < 0 {
		return nil, model.ErrValidation
	}

	o, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items, err := s.orders.GetItemsByIDs(ctx, orderID, itemIDs)
	if err != nil {
		return nil, err
	}
	if len(items) != len(itemIDs) {
		return nil, ErrItemNotInOrder
	}
	amount := 0
	for _, it := range items {
		if it.Status != ordermodel.ItemStatusUnpaid {
			return nil, ErrItemAlreadyPaid
		}
		amount += it.Price
	}
	amount += tipsAmount

	t, err := s.tables.GetByID(ctx, o.TableID)
	if err != nil {
		return nil, err
	}

	pool := s.orders.Pool()
	dbtx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = dbtx.Rollback(ctx)
		}
	}()

	txn := &tx_model.Transaction{
		ID:         uuid.New().String(),
		OrderID:    orderID,
		Amount:     amount,
		TipsAmount: tipsAmount,
		Status:     tx_model.StatusPending,
		ItemIDs:    itemIDs,
		CreatedAt:  time.Now().UTC(),
	}
	if err = s.txRepo.Create(ctx, dbtx, txn); err != nil {
		return nil, err
	}
	if err = s.txRepo.LinkItems(ctx, dbtx, txn.ID, itemIDs); err != nil {
		return nil, err
	}

	chargeStatus, chargeErr := s.pay.Charge(ctx, txn.ID, amount)
	if chargeErr != nil || chargeStatus != tx_model.StatusSuccess {
		if uerr := s.txRepo.UpdateStatus(ctx, dbtx, txn.ID, tx_model.StatusFailed); uerr != nil {
			err = uerr
			return nil, err
		}
		if cerr := dbtx.Commit(ctx); cerr != nil {
			err = cerr
			return nil, err
		}
		err = ErrPaymentFailed
		return nil, err
	}

	if err = s.orders.MarkItemsPaid(ctx, dbtx, itemIDs); err != nil {
		return nil, err
	}
	newPaid := o.PaidAmount
	for _, it := range items {
		newPaid += it.Price
	}
	if err = s.orders.UpdateTotals(ctx, dbtx, orderID, o.TotalAmount, newPaid); err != nil {
		return nil, err
	}

	if err = s.txRepo.UpdateStatus(ctx, dbtx, txn.ID, tx_model.StatusSuccess); err != nil {
		return nil, err
	}
	txn.Status = tx_model.StatusSuccess

	if err = dbtx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	if newPaid >= o.TotalAmount {
		t.Status = tablemodel.StatusPaid
	} else {
		t.Status = tablemodel.StatusUnpaid
	}
	if err = s.tables.Update(ctx, t); err != nil {
		return nil, err
	}

	return txn, nil
}
