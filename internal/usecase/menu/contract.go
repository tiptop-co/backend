package menu

import (
	"context"
	"errors"

	"github.com/tiptop-co/backend/internal/model/dish"
)

var (
	ErrInvalidWeightUnit = errors.New("invalid weight unit")
	ErrCategoryNotInVenue = errors.New("category does not belong to venue")
)

type MenuUsecase interface {
	GetVenueMenu(ctx context.Context, venueID string) ([]*dish.Dish, []*dish.Category, error)
	GetDish(ctx context.Context, id string) (*dish.Dish, error)
	CreateDish(ctx context.Context, venueID string, in *dish.CreateInput) (*dish.Dish, error)
	DeleteDish(ctx context.Context, venueID, dishID string) error
}

type MenuRepository interface {
	GetCategoriesByVenue(ctx context.Context, venueID string) ([]*dish.Category, error)
	GetDishesByVenue(ctx context.Context, venueID string) ([]*dish.Dish, error)
	GetDishByID(ctx context.Context, id string) (*dish.Dish, error)
	CategoryBelongsToVenue(ctx context.Context, categoryID, venueID string) (bool, error)
	CreateDish(ctx context.Context, d *dish.Dish) error
	DeleteDish(ctx context.Context, dishID, venueID string) error
}
