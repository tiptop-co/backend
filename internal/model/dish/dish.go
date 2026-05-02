package dish

type WeightUnit string

const (
	WeightUnitGram WeightUnit = "г"
	WeightUnitMl   WeightUnit = "мл"
)

func (u WeightUnit) IsValid() bool {
	return u == WeightUnitGram || u == WeightUnitMl
}

type Category struct {
	ID      string `json:"id"       db:"id"`
	Name    string `json:"name"     db:"name"`
	VenueID string `json:"venue_id" db:"venue_id"`
}

type Dish struct {
	ID           string     `json:"id"            db:"id"`
	Name         string     `json:"name"          db:"name"`
	Description  string     `json:"description"   db:"description"`
	CategoryID   string     `json:"category_id"   db:"category_id"`
	CategoryName string     `json:"category_name" db:"category_name"`
	Price        int        `json:"price"         db:"price"`
	Weight       int        `json:"weight"        db:"weight"`
	WeightUnit   WeightUnit `json:"weight_unit"   db:"weight_unit"`
	Calories     *int       `json:"calories"      db:"calories"`
	Protein      *int       `json:"protein"       db:"protein"`
	Fat          *int       `json:"fat"           db:"fat"`
	Carbs        *int       `json:"carbs"         db:"carbs"`
	VenueID      string     `json:"venue_id"      db:"venue_id"`
}

type CreateInput struct {
	Name        string
	Description string
	CategoryID  string
	Price       int
	Weight      int
	WeightUnit  WeightUnit
	Calories    *int
	Protein     *int
	Fat         *int
	Carbs       *int
}
