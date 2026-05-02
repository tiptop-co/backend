package manager_handler

import "github.com/tiptop-co/backend/internal/model/dish"

type createDishRequest struct {
	Name        string  `json:"name"        binding:"required"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id" binding:"required"`
	Price       int     `json:"price"       binding:"required"`
	Weight      int     `json:"weight"      binding:"required"`
	WeightUnit  string  `json:"weight_unit" binding:"required"`
	Calories    *int    `json:"calories"`
	Protein     *int    `json:"protein"`
	Fat         *int    `json:"fat"`
	Carbs       *int    `json:"carbs"`
}

type getMenuResponse struct {
	Dishes     []*dish.Dish     `json:"dishes"`
	Categories []*dish.Category `json:"categories"`
}
