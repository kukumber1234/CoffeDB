package models

import "time"

type MenuItem struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Tags        []string `json:"tags"`
}

type MenuItemIngredient struct {
	ID           int     `json:"id"`
	MenuItemID   int     `json:"menu_item_id"`
	IngredientID int     `json:"ingredient_id"`
	Quantity     float64 `json:"quantity"`
}

type MenuPriceHistory struct {
	ID         int       `json:"id"`
	MenuItemID int       `json:"menu_item_id"`
	OldPrice   float64   `json:"old_price"`
	NewPrice   float64   `json:"new_price"`
	ChangedAt  time.Time `json:"changed_at"`
}

type MenuResponse struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Price       float64              `json:"price"`
	Ingredients []MenuItemIngredient `json:"ingredients"`
}

type MenuRequest struct {
	Menu            MenuItem        `json:"menu_item"`
	MenuIngredients []MenuInventory `json:"ingredients"`
}

type MenuInventory struct {
	Inventory InventoryMenuRequest `json:"inventory"`
	Quantity  float64              `json:"quantity"`
}
