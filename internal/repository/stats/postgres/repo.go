package stats_postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tiptop-co/backend/internal/model/stats"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func todayBounds() (time.Time, time.Time) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return start, start.Add(24 * time.Hour)
}

func (r *Repository) VenueStats(ctx context.Context, venueID string) (*stats.VenueStats, error) {
	dayStart, dayEnd := todayBounds()

	out := &stats.VenueStats{}

	const aggQuery = `
		SELECT COALESCE(SUM(t.amount), 0) AS revenue,
		       COUNT(DISTINCT t.order_id) AS orders_count,
		       COALESCE(SUM(t.tips_amount), 0) AS tips_total
		FROM transactions t
		JOIN orders o ON o.id = t.order_id
		JOIN tables tb ON tb.id = o.table_id
		WHERE t.status = 'success'
		  AND tb.venue_id = $1
		  AND t.created_at >= $2 AND t.created_at < $3
	`
	if err := r.db.QueryRow(ctx, aggQuery, venueID, dayStart, dayEnd).Scan(&out.Revenue, &out.OrdersCount, &out.TipsTotal); err != nil {
		return nil, fmt.Errorf("stats venue agg: %w", err)
	}
	if out.OrdersCount > 0 {
		out.AverageCheck = out.Revenue / out.OrdersCount
	}

	const dailyQuery = `
		SELECT to_char(d.day, 'YYYY-MM-DD') AS day,
		       COALESCE(SUM(t.amount), 0)
		FROM generate_series($2::date, $3::date, '1 day') AS d(day)
		LEFT JOIN transactions t ON t.status = 'success'
		    AND date_trunc('day', t.created_at) = d.day
		LEFT JOIN orders o ON o.id = t.order_id
		LEFT JOIN tables tb ON tb.id = o.table_id AND tb.venue_id = $1
		WHERE t.id IS NULL OR tb.venue_id = $1
		GROUP BY d.day
		ORDER BY d.day
	`
	weekStart := dayStart.AddDate(0, 0, -6)
	rows, err := r.db.Query(ctx, dailyQuery, venueID, weekStart, dayStart)
	if err != nil {
		return nil, fmt.Errorf("stats venue daily: %w", err)
	}
	for rows.Next() {
		var d stats.DailyRevenue
		if err := rows.Scan(&d.Day, &d.Amount); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan daily: %w", err)
		}
		out.DailyRevenue = append(out.DailyRevenue, &d)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("daily rows: %w", err)
	}

	const topQuery = `
		SELECT oi.dish_name, SUM(oi.quantity)::int AS cnt
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN tables tb ON tb.id = o.table_id
		WHERE tb.venue_id = $1
		  AND o.created_at >= $2 AND o.created_at < $3
		GROUP BY oi.dish_name
		ORDER BY cnt DESC
		LIMIT 5
	`
	trows, err := r.db.Query(ctx, topQuery, venueID, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("stats venue top: %w", err)
	}
	for trows.Next() {
		var d stats.TopDish
		if err := trows.Scan(&d.Name, &d.Count); err != nil {
			trows.Close()
			return nil, fmt.Errorf("scan top: %w", err)
		}
		out.TopDishes = append(out.TopDishes, &d)
	}
	trows.Close()
	if err := trows.Err(); err != nil {
		return nil, fmt.Errorf("top rows: %w", err)
	}

	return out, nil
}

func (r *Repository) GlobalStats(ctx context.Context) (*stats.GlobalStats, error) {
	dayStart, dayEnd := todayBounds()

	out := &stats.GlobalStats{}

	const aggQuery = `
		SELECT
		  (SELECT COUNT(*) FROM venues),
		  COALESCE(SUM(t.amount), 0),
		  COUNT(DISTINCT t.order_id)
		FROM transactions t
		WHERE t.status = 'success'
		  AND t.created_at >= $1 AND t.created_at < $2
	`
	if err := r.db.QueryRow(ctx, aggQuery, dayStart, dayEnd).Scan(&out.VenuesCount, &out.TotalRevenue, &out.TotalOrders); err != nil {
		return nil, fmt.Errorf("stats global agg: %w", err)
	}
	if out.TotalOrders > 0 {
		out.AverageCheck = out.TotalRevenue / out.TotalOrders
	}

	const perVenueQuery = `
		SELECT v.id, v.name,
		       COALESCE(SUM(t.amount), 0) AS revenue,
		       COUNT(DISTINCT t.order_id) AS orders
		FROM venues v
		LEFT JOIN tables tb ON tb.venue_id = v.id
		LEFT JOIN orders o ON o.table_id = tb.id
		LEFT JOIN transactions t ON t.order_id = o.id AND t.status = 'success'
		    AND t.created_at >= $1 AND t.created_at < $2
		GROUP BY v.id, v.name
		ORDER BY revenue DESC
	`
	rows, err := r.db.Query(ctx, perVenueQuery, dayStart, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("stats global per-venue: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v stats.VenueAggregate
		if err := rows.Scan(&v.VenueID, &v.Name, &v.Revenue, &v.Orders); err != nil {
			return nil, fmt.Errorf("scan per-venue: %w", err)
		}
		if v.Orders > 0 {
			v.AverageCheck = v.Revenue / v.Orders
		}
		out.Venues = append(out.Venues, &v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("per-venue rows: %w", err)
	}

	return out, nil
}
