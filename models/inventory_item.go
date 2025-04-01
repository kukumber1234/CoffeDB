package models

import "time"

type InventoryItem struct {
	IngredientID int       `json:"ingredient_id"`
	Name         string    `json:"name"`
	StockLevel   float64   `json:"stock_level"`
	Unit         string    `json:"unit"`
	LastUpdated  time.Time `json:"last_updated"`
}

type InventoryMenuRequest struct {
	Name string `json:"name"`
}

type CountInventory struct {
	CurrentPage int    `json:"currentPage"`
	HasNextPage bool   `json:"hasNextPage"`
	PageSize    int    `json:"pageSize"`
	TotalPages  int    `json:"totalPages"`
	Data        []Data `json:"data"`
}

type Data struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int    `json:"price"`
}
