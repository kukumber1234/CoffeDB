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

type BatchOrderRequest struct {
	Orders []OrderRequestBatch `json:"orders"`
}

type OrderRequestBatch struct {
	CustomerName string                  `json:"customer_name"`
	Items        []OrderItemRequestBatch `json:"items"`
}

type OrderItemRequestBatch struct {
	MenuItemName string `json:"product_name"`
	Quantity     int    `json:"quantity"`
}

type ProcessedOrder struct {
	OrderID      int     `json:"order_id,omitempty"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Total        float64 `json:"total,omitempty"`
	Reason       string  `json:"reason,omitempty"`
}

type InventoryUpdate struct {
	IngredientID int     `json:"ingredient_id"`
	Name         string  `json:"name"`
	QuantityUsed float64 `json:"quantity_used"`
	Remaining    float64 `json:"remaining"`
}

type BatchSummary struct {
	TotalOrders      int               `json:"total_orders"`
	Accepted         int               `json:"accepted"`
	Rejected         int               `json:"rejected"`
	TotalRevenue     float64           `json:"total_revenue"`
	InventoryUpdates []InventoryUpdate `json:"inventory_updates"`
}

type BatchOrderResponse struct {
	ProcessedOrders []ProcessedOrder `json:"processed_orders"`
	Summary         BatchSummary     `json:"summary"`
}
