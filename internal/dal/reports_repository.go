package dal

import (
	"database/sql"
	model "frapo/models"
	"strconv"

	"github.com/lib/pq"
)

type ReportsDalInterface interface {
	TotalPrice() (model.TotalSalesStruct, error)
	PopularItems(limit string) ([]model.PopularItem, error)
	FullTextSearchMenu(q, minPrice, maxPrice string) (int, []model.MenuItemResult, error)
	FullTextSearchOrder(q, minPrice, maxPrice string) (int, []model.OrderResult, error)
	OrderedItemsByPeriodDay(month int) (model.ItemByPeriodMonth, error)
	OrderedItemsByPeriodMonth(year int) (model.ItemByPeriodYear, error)
}

type ReportsData struct {
	db *sql.DB
}

func NewReportsRepo(db *sql.DB) *ReportsData {
	return &ReportsData{db: db}
}

func (f *ReportsData) TotalPrice() (model.TotalSalesStruct, error) {
	query := `SELECT COALESCE(SUM(total_amount), 0) FROM orders
			  WHERE status = 'closed'`

	var totalPrice model.TotalSalesStruct
	if err := f.db.QueryRow(query).Scan(&totalPrice.TotalSales); err != nil {
		return model.TotalSalesStruct{}, err
	}

	return totalPrice, nil
}

func (f *ReportsData) PopularItems(limit string) ([]model.PopularItem, error) {
	query := `SELECT mi.name, SUM(oi.quantity) AS totalQuant
			  FROM menu_items mi
			  JOIN order_items oi ON mi.menu_item_id = oi.menu_item_id
			  JOIN orders o ON oi.order_id = o.order_id
			  WHERE o.status = 'closed'
			  GROUP BY mi.name
			  ORDER BY totalQuant DESC
			  LIMIT $1
		`

	rows, err := f.db.Query(query, limit)
	if err != nil {
		return nil, err
	}

	var items []model.PopularItem

	for rows.Next() {
		var item model.PopularItem

		if err := rows.Scan(&item.Name, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (f *ReportsData) FullTextSearchMenu(q, minPrice, maxPrice string) (int, []model.MenuItemResult, error) {
	var menuItems []model.MenuItemResult

	query := `
		SELECT 
			menu_item_id AS id,
			name,
			description,
			price,
			ts_rank(
				to_tsvector('english', name) || to_tsvector('english', description),
				plainto_tsquery('english', $1)
			) AS relevance
		FROM menu_items
		WHERE 
			(to_tsvector('english', name) @@ plainto_tsquery('english', $1)
			 OR to_tsvector('english', description) @@ plainto_tsquery('english', $1))
			AND ($2::NUMERIC IS NULL OR price >= $2::NUMERIC)
			AND ($3::NUMERIC IS NULL OR price <= $3::NUMERIC)
		ORDER BY relevance DESC;
	`

	rows, err := f.db.Query(query, q, minPrice, maxPrice)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var menuItem model.MenuItemResult
		if err := rows.Scan(&menuItem.ID, &menuItem.Name, &menuItem.Description, &menuItem.Price, &menuItem.Relevance); err != nil {
			return 0, nil, err
		}
		menuItems = append(menuItems, menuItem)
	}
	totalMatches := len(menuItems)
	return totalMatches, menuItems, nil
}

func (f *ReportsData) FullTextSearchOrder(q, minPrice, maxPrice string) (int, []model.OrderResult, error) {
	var orderResults []model.OrderResult

	query := `
		SELECT 
			o.order_id AS id,
			o.customer_name,
			array_agg(mi.name) AS items,
			o.total_amount AS total,
			ts_rank(
				setweight(to_tsvector('english', o.customer_name), 'A') || 
				setweight(to_tsvector('english', string_agg(mi.name, ' ')), 'B'),
				plainto_tsquery('english', $1)
			) AS relevance
		FROM orders o
		JOIN order_items oi ON o.order_id = oi.order_id
		JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
		WHERE 
			(to_tsvector('english', o.customer_name) @@ plainto_tsquery('english', $1)
			 OR to_tsvector('english', mi.name) @@ plainto_tsquery('english', $1))
			AND ($2::NUMERIC IS NULL OR o.total_amount >= $2::NUMERIC)
			AND ($3::NUMERIC IS NULL OR o.total_amount <= $3::NUMERIC)
		GROUP BY o.order_id, o.customer_name, o.total_amount
		ORDER BY relevance DESC;
	`

	rows, err := f.db.Query(query, q, minPrice, maxPrice)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var orderResult model.OrderResult
		// Важно: порядок сканирования соответствует порядку SELECT!
		if err := rows.Scan(&orderResult.ID, &orderResult.Customer_name, pq.Array(&orderResult.Items), &orderResult.Total, &orderResult.Relevance); err != nil {
			return 0, nil, err
		}
		orderResults = append(orderResults, orderResult)
	}
	totalMatches := len(orderResults)
	return totalMatches, orderResults, nil
}

func (f *ReportsData) OrderedItemsByPeriodDay(month int) (model.ItemByPeriodMonth, error) {
	query := `SELECT EXTRACT(DAY FROM order_date) AS day, COUNT(*) AS orders
			  FROM orders
			  WHERE EXTRACT(MONTH FROM order_date) = $1
			  GROUP BY day
			  ORDER BY day`

	rows, err := f.db.Query(query, month)
	if err != nil {
		return model.ItemByPeriodMonth{}, err
	}

	var orderItems []model.OrderDayCount

	for rows.Next() {
		var orderItem model.OrderDayCount
		if err := rows.Scan(&orderItem.Day, &orderItem.Count); err != nil {
			return model.ItemByPeriodMonth{}, err
		}
		orderItems = append(orderItems, orderItem)
	}

	itemPeriodMonth := model.ItemByPeriodMonth{
		Period:       "day",
		Month:        monthToString(month),
		OrderedItems: orderItems,
	}

	return itemPeriodMonth, nil
}

func (f *ReportsData) OrderedItemsByPeriodMonth(year int) (model.ItemByPeriodYear, error) {
	query := `SELECT EXTRACT(MONTH FROM order_date) AS month, COUNT(*) AS orders
			  FROM orders
			  WHERE EXTRACT(YEAR FROM order_date) = $1
			  GROUP BY month
			  ORDER BY month`

	rows, err := f.db.Query(query, year)
	if err != nil {
		return model.ItemByPeriodYear{}, err
	}

	var orderedItems []model.OrderMonthCount

	for rows.Next() {
		var count, month int
		if err := rows.Scan(&month, &count); err != nil {
			return model.ItemByPeriodYear{}, err
		}
		orderedItems = append(orderedItems, model.OrderMonthCount{
			Month: monthToString(month),
			Count: count,
		})
	}

	itemByPeriodYear := model.ItemByPeriodYear {
		Period: "month",
		Year: strconv.Itoa(year),
		OrderedItems: orderedItems,
	}

	return itemByPeriodYear, nil
}

func monthToString(m int) string {
	months := map[int]string{
		1:  "january",
		2:  "february",
		3:  "march",
		4:  "april",
		5:  "may",
		6:  "june",
		7:  "july",
		8:  "august",
		9:  "september",
		10: "october",
		11: "november",
		12: "december",
	}
	return months[m]
}