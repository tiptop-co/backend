package menu_postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model"
	"github.com/tiptop-co/backend/internal/model/dish"
)

type MenuRepository struct {
	db *pgxpool.Pool
}

func NewMenuRepository(db *pgxpool.Pool) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) GetCategoriesByVenue(ctx context.Context, venueID string) ([]*dish.Category, error) {
	const query = `SELECT id, name, venue_id FROM menu_categories WHERE venue_id = $1 ORDER BY name`

	rows, err := r.db.Query(ctx, query, venueID)
	if err != nil {
		return nil, fmt.Errorf("menu repository get categories: %w", err)
	}
	defer rows.Close()

	var result []*dish.Category
	for rows.Next() {
		var c dish.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.VenueID); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		result = append(result, &c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *MenuRepository) GetDishesByVenue(ctx context.Context, venueID string) ([]*dish.Dish, error) {
	const query = `
		SELECT d.id, d.name, COALESCE(d.description,'') AS description,
		       d.category_id, c.name AS category_name,
		       d.price, d.weight, d.weight_unit,
		       d.calories, d.protein, d.fat, d.carbs, d.venue_id
		FROM dishes d
		LEFT JOIN menu_categories c ON c.id = d.category_id
		WHERE d.venue_id = $1
		ORDER BY d.name
	`
	rows, err := r.db.Query(ctx, query, venueID)
	if err != nil {
		return nil, fmt.Errorf("menu repository get dishes: %w", err)
	}
	defer rows.Close()

	var result []*dish.Dish
	for rows.Next() {
		d, err := scanDish(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("rows: %w", rows.Err())
	}
	return result, nil
}

func (r *MenuRepository) GetDishByID(ctx context.Context, id string) (*dish.Dish, error) {
	const query = `
		SELECT d.id, d.name, COALESCE(d.description,'') AS description,
		       d.category_id, c.name AS category_name,
		       d.price, d.weight, d.weight_unit,
		       d.calories, d.protein, d.fat, d.carbs, d.venue_id
		FROM dishes d
		LEFT JOIN menu_categories c ON c.id = d.category_id
		WHERE d.id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	d, err := scanDish(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("menu repository get dish by id: %w", err)
	}
	return d, nil
}

func (r *MenuRepository) CategoryBelongsToVenue(ctx context.Context, categoryID, venueID string) (bool, error) {
	const query = `SELECT 1 FROM menu_categories WHERE id = $1 AND venue_id = $2`
	var x int
	err := r.db.QueryRow(ctx, query, categoryID, venueID).Scan(&x)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("menu repository category belongs: %w", err)
	}
	return true, nil
}

func (r *MenuRepository) CreateDish(ctx context.Context, d *dish.Dish) error {
	const query = `
		INSERT INTO dishes (id, name, description, category_id, price, weight, weight_unit,
		                    calories, protein, fat, carbs, venue_id)
		VALUES (@id, @name, @description, @category_id, @price, @weight, @weight_unit,
		        @calories, @protein, @fat, @carbs, @venue_id)
	`
	args := pgx.NamedArgs{
		"id":          d.ID,
		"name":        d.Name,
		"description": d.Description,
		"category_id": d.CategoryID,
		"price":       d.Price,
		"weight":      d.Weight,
		"weight_unit": string(d.WeightUnit),
		"calories":    d.Calories,
		"protein":     d.Protein,
		"fat":         d.Fat,
		"carbs":       d.Carbs,
		"venue_id":    d.VenueID,
	}
	if _, err := r.db.Exec(ctx, query, args); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			err = fmt.Errorf("%w: %w", model.ErrAlreadyExists, err)
		}
		return fmt.Errorf("menu repository create dish: %w", err)
	}
	return nil
}

func (r *MenuRepository) DeleteDish(ctx context.Context, dishID, venueID string) error {
	const query = `DELETE FROM dishes WHERE id = $1 AND venue_id = $2`
	cmd, err := r.db.Exec(ctx, query, dishID, venueID)
	if err != nil {
		return fmt.Errorf("menu repository delete dish: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDish(s rowScanner) (*dish.Dish, error) {
	var (
		d            dish.Dish
		categoryName *string
		weightUnit   string
	)
	err := s.Scan(
		&d.ID, &d.Name, &d.Description,
		&d.CategoryID, &categoryName,
		&d.Price, &d.Weight, &weightUnit,
		&d.Calories, &d.Protein, &d.Fat, &d.Carbs, &d.VenueID,
	)
	if err != nil {
		return nil, err
	}
	d.WeightUnit = dish.WeightUnit(weightUnit)
	if categoryName != nil {
		d.CategoryName = *categoryName
	}
	return &d, nil
}
