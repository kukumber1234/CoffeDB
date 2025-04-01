package models

import "time"

type Order struct {
	OrderID            int                    `json:"order_id"`
	CustomerName       string                 `json:"customer_name"`
	OrderDate          time.Time              `json:"order_date"`
	Status             string                 `json:"status"`
	TotalPrice         float64                `json:"total_amount"`
	SpecialInstruction map[string]interface{} `json:"special_instruction"`
}

type OrderItem struct {
	OrderItemID      int                    `json:"product_id"`
	MenuItemID       int                    `json:"menu_item_id"`
	OrderID          int                    `json:"order_id"`
	Customizations   map[string]interface{} `json:"customizations"`
	PriceAtOrderTime float64                `json:"price_at_order_time"`
	Quantity         int                    `json:"quantity"`
}

type OrderStatusHistory struct {
	ID       int    `json:"id"`
	OrderID  int    `json:"order_id"`
	Status   string `json:"status"`
	ChangeAt string `json:"change_at"`
}

type OrderRequest struct {
	CustomerName string             `json:"customer_name"`
	Orders       []OrderItemRequest `json:"orders"`
}

type OrderItemRequest struct {
	MenuItemID string `json:"product_id"`
	Quantity   int    `json:"quantity"`
}

type InventoryUpdates struct {
	IngredientID int    `json:"ingredient_id"`
	Name         string `json:"name"`
	QuantityUsed int    `json:"quantity_used"`
	Remaining    int    `json:"remaining"`
}

type OrderResponse struct {
	OrderID      int              `json:"order_id"`
	CustomerName string           `json:"customer_name"`
	Items        []OrderItemShort `json:"items"`
	Status       string           `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
}

type OrderItemShort struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type NumberOfOrderedItemsResponse map[string]int
