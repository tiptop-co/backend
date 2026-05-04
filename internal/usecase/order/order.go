package order

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/tiptop-co/backend/internal/model"
	ordermodel "github.com/tiptop-co/backend/internal/model/order"
	tablemodel "github.com/tiptop-co/backend/internal/model/table"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[ORDER USECASE]"

type OrderService struct {
	orders   OrderRepository
	dishes   DishLookup
	tables   TableRepo
	assigner WaiterAssigner
}

func NewOrderService(orders OrderRepository, dishes DishLookup, tables TableRepo, assigner WaiterAssigner) *OrderService {
	return &OrderService{orders: orders, dishes: dishes, tables: tables, assigner: assigner}
}

func (s *OrderService) GetCompletedByWaiter(ctx context.Context, waiterID string, limit int) (_ []*ordermodel.CompletedSummary, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get completed by waiter", err) }()

	if waiterID == "" {
		return nil, model.ErrValidation
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.orders.GetCompletedByWaiter(ctx, waiterID, limit)
}

func (s *OrderService) GetByTable(ctx context.Context, tableID string) (_ *ordermodel.Order, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get by table", err) }()

	if tableID == "" {
		return nil, model.ErrValidation
	}
	return s.orders.GetActiveByTable(ctx, tableID)
}

func (s *OrderService) Create(ctx context.Context, tableID string, items []ordermodel.CreateItemInput, wishes *string) (_ *ordermodel.Order, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create", err) }()

	if tableID == "" {
		return nil, model.ErrValidation
	}
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}
	for _, it := range items {
		if it.DishID == "" || it.Quantity <= 0 {
			return nil, model.ErrValidation
		}
	}

	t, err := s.tables.GetByID(ctx, tableID)
	if err != nil {
		return nil, err
	}

	resolved := make([]ordermodel.Item, 0, len(items))
	addedSum := 0
	for _, in := range items {
		d, err := s.dishes.GetDishByID(ctx, in.DishID)
		if err != nil {
			return nil, err
		}
		if d.VenueID != t.VenueID {
			return nil, ErrDishWrongVenue
		}
		resolved = append(resolved, ordermodel.Item{
			DishID:   d.ID,
			DishName: d.Name,
			Quantity: in.Quantity,
			Price:    d.Price * in.Quantity,
			Status:   ordermodel.ItemStatusUnpaid,
		})
		addedSum += d.Price * in.Quantity
	}

	pool := s.orders.Pool()
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	existing, err := s.orders.GetActiveByTable(ctx, tableID)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}

	var orderID string

	if existing != nil {
		orderID = existing.ID

		for i := range resolved {
			resolved[i].ID = uuid.New().String()
			resolved[i].OrderID = orderID
			resolved[i].AddedLater = true
		}
		if err = s.orders.AddItems(ctx, tx, resolved); err != nil {
			return nil, err
		}
		if err = s.orders.UpdateTotals(ctx, tx, orderID, existing.TotalAmount+addedSum, existing.PaidAmount); err != nil {
			return nil, err
		}
	} else {
		orderID = uuid.New().String()
		o := &ordermodel.Order{
			ID:          orderID,
			TableID:     tableID,
			WaiterID:    t.WaiterID,
			Status:      ordermodel.StatusActive,
			TotalAmount: addedSum,
			PaidAmount:  0,
			Wishes:      wishes,
			CreatedAt:   time.Now().UTC(),
		}
		if err = s.orders.Create(ctx, tx, o); err != nil {
			return nil, err
		}
		for i := range resolved {
			resolved[i].ID = uuid.New().String()
			resolved[i].OrderID = orderID
			resolved[i].AddedLater = false
		}
		if err = s.orders.AddItems(ctx, tx, resolved); err != nil {
			return nil, err
		}
	}

	if t.WaiterID == nil {
		waiterID, err2 := s.assigner.LeastLoadedWaiter(ctx, t.VenueID)
		if err2 != nil {
			err = err2
			return nil, err
		}
		if waiterID != nil {
			t.WaiterID = waiterID
			if err = s.orders.UpdateWaiter(ctx, tx, orderID, waiterID); err != nil {
				return nil, err
			}
		}
	}

	t.OrderID = &orderID
	if t.Status == tablemodel.StatusFree {
		t.Status = tablemodel.StatusUnpaid
	}
	if err = s.tables.Update(ctx, t); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return s.orders.GetByID(ctx, orderID)
}
