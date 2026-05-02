package waiter_request

import (
	"context"
	"errors"

	"github.com/tiptop-co/backend/internal/model/table"
	wrmodel "github.com/tiptop-co/backend/internal/model/waiter_request"
)

var (
	ErrCallNotAllowed = errors.New("call not allowed: previous request still active")
	ErrNotOwnRequest  = errors.New("request belongs to other waiter")
)

type Usecase interface {
	CanCall(ctx context.Context, tableID string) (bool, error)
	GetByTable(ctx context.Context, tableID string) ([]*wrmodel.Request, error)
	GetByWaiter(ctx context.Context, waiterID string) ([]*wrmodel.Request, error)
	Create(ctx context.Context, tableID string) (*wrmodel.Request, error)
	Accept(ctx context.Context, requestID, waiterID string) (*wrmodel.Request, error)
}

type Repository interface {
	Create(ctx context.Context, req *wrmodel.Request) error
	GetByID(ctx context.Context, id string) (*wrmodel.Request, error)
	GetByFilters(ctx context.Context, f *wrmodel.Filters) ([]*wrmodel.Request, error)
	UpdateStatus(ctx context.Context, id string, status wrmodel.Status) error
	UpdateWaiter(ctx context.Context, id string, waiterID string) error
}

type TableLookup interface {
	GetByID(ctx context.Context, id string) (*table.Table, error)
}

type WaiterAssigner interface {
	LeastLoadedWaiter(ctx context.Context, venueID string) (*string, error)
}
