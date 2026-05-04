package menu

import (
	"context"

	"github.com/google/uuid"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/dish"
	"github.com/tiptop-co/backend/pkg/errwrap"
)

const errPrefix = "[MENU USECASE]"

type MenuService struct {
	repo MenuRepository
}

func NewMenuService(repo MenuRepository) *MenuService {
	return &MenuService{repo: repo}
}

func (s *MenuService) GetVenueMenu(ctx context.Context, venueID string) (_ []*dish.Dish, _ []*dish.Category, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get venue menu", err) }()

	if venueID == "" {
		return nil, nil, model.ErrValidation
	}

	dishes, err := s.repo.GetDishesByVenue(ctx, venueID)
	if err != nil {
		return nil, nil, err
	}
	categories, err := s.repo.GetCategoriesByVenue(ctx, venueID)
	if err != nil {
		return nil, nil, err
	}
	return dishes, categories, nil
}

func (s *MenuService) GetDish(ctx context.Context, id string) (_ *dish.Dish, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" get dish", err) }()

	if id == "" {
		return nil, model.ErrValidation
	}
	return s.repo.GetDishByID(ctx, id)
}

func (s *MenuService) CreateDish(ctx context.Context, venueID string, in *dish.CreateInput) (_ *dish.Dish, err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" create dish", err) }()

	if in == nil || venueID == "" || in.Name == "" || in.CategoryID == "" {
		return nil, model.ErrValidation
	}
	if in.Price <= 0 || in.Weight <= 0 {
		return nil, model.ErrValidation
	}
	if !in.WeightUnit.IsValid() {
		return nil, ErrInvalidWeightUnit
	}

	ok, err := s.repo.CategoryBelongsToVenue(ctx, in.CategoryID, venueID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrCategoryNotInVenue
	}

	d := &dish.Dish{
		ID:          uuid.New().String(),
		Name:        in.Name,
		Description: in.Description,
		CategoryID:  in.CategoryID,
		Price:       in.Price,
		Weight:      in.Weight,
		WeightUnit:  in.WeightUnit,
		Calories:    in.Calories,
		Protein:     in.Protein,
		Fat:         in.Fat,
		Carbs:       in.Carbs,
		VenueID:     venueID,
	}
	if err := s.repo.CreateDish(ctx, d); err != nil {
		return nil, err
	}

	full, err := s.repo.GetDishByID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	return full, nil
}

func (s *MenuService) DeleteDish(ctx context.Context, venueID, dishID string) (err error) {
	defer func() { err = errwrap.WrapMsg(errPrefix+" delete dish", err) }()

	if venueID == "" || dishID == "" {
		return model.ErrValidation
	}
	return s.repo.DeleteDish(ctx, dishID, venueID)
}
