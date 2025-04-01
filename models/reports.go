package models

type TotalSalesStruct struct {
	TotalSales float64 `json:"total_sales"`
}

type PopularItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type SearchResponse struct {
	MenuItems    []MenuItemResult `json:"menu_items,omitempty"`
	Orders       []OrderResult    `json:"orders,omitempty"`
	TotalMatches int              `json:"total_matches"`
}

type MenuItemResult struct {
	ID          string  `json:"ID"`
	Name        string  `json:"Name"`
	Description string  `json:"Description"`
	Price       float64 `json:"Price"`
	Relevance   float64 `json:"Relevance"`
}

type OrderResult struct {
	ID            string   `json:"ID"`
	Customer_name string   `json:"Customer_name"`
	Items         []string `json:"Items"`
	Total         float64  `json:"Total"`
	Relevance     float64  `json:"Relevance"`
}

type ItemByPeriodMonth struct {
	Period       string          `json:"period"`
	Month        string          `json:"month"`
	OrderedItems []OrderDayCount `json:"orderedItems"`
}

type OrderDayCount struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type ItemByPeriodYear struct {
	Period       string            `json:"period"`
	Year         string            `json:"year"`
	OrderedItems []OrderMonthCount `json:"orderedItems"`
}

type OrderMonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}
