package table_close

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/tiptop-co/backend/internal/model"
	ordermodel "github.com/tiptop-co/backend/internal/model/order"
	tablemodel "github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[TABLE_CLOSE USECASE]"

type Service struct {
	tables       TableRepo
	orders       OrderRepo
	gen          TokenGenerator
	sessionSize  int
}

func NewService(tables TableRepo, orders OrderRepo, gen TokenGenerator, sessionSize int) *Service {
	return &Service{tables: tables, orders: orders, gen: gen, sessionSize: sessionSize}
}

func (s *Service) Close(ctx context.Context, tableID string) (_ *tablemodel.Table, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" close", err) }()

	if tableID == "" {
		return nil, model.ErrValidation
	}

	t, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, err
	}

	o, err := s.orders.GetActiveByTable(ctx, tableID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, err
	}

	if o != nil {
		for _, it := range o.Items {
			if it.Status != ordermodel.ItemStatusPaid {
				return nil, ErrUnpaidItems
			}
		}
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

	if o != nil {
		if err = s.orders.MarkCompleted(ctx, dbtx, o.ID); err != nil {
			return nil, err
		}
	}
	if err = dbtx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	newSession, err := s.gen.GenerateToken(s.sessionSize)
	if err != nil {
		return nil, fmt.Errorf("generate session token: %w", err)
	}

	t.Status = tablemodel.StatusFree
	t.WaiterID = nil
	t.OrderID = nil
	t.SessionToken = newSession

	if err = s.tables.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}
