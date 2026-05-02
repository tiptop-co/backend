package stats

type DailyRevenue struct {
	Day    string `json:"day"`
	Amount int    `json:"amount"`
}

type TopDish struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type VenueAggregate struct {
	VenueID      string `json:"venue_id,omitempty"`
	Name         string `json:"name"`
	Revenue      int    `json:"revenue"`
	Orders       int    `json:"orders"`
	AverageCheck int    `json:"average_check"`
}

type VenueStats struct {
	Revenue      int             `json:"revenue"`
	OrdersCount  int             `json:"orders_count"`
	AverageCheck int             `json:"average_check"`
	TipsTotal    int             `json:"tips_total"`
	DailyRevenue []*DailyRevenue `json:"daily_revenue"`
	TopDishes    []*TopDish      `json:"top_dishes"`
}

type GlobalStats struct {
	VenuesCount  int               `json:"venues_count"`
	TotalRevenue int               `json:"total_revenue"`
	TotalOrders  int               `json:"total_orders"`
	AverageCheck int               `json:"average_check"`
	Venues       []*VenueAggregate `json:"venues"`
}
