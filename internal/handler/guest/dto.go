package guest_handler

import (
	"github.com/tiptop-co/backend/internal/model/dish"
)

type getMenuResponse struct {
	Dishes     []*dish.Dish     `json:"dishes"`
	Categories []*dish.Category `json:"categories"`
}
