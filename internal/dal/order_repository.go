package dal

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	model "frappuccino/models"
)

type OrderRepository interface {
	Add(name string, itemReq []model.OrderItemRequest) (int, []model.InventoryUpdate, error)
	GetAll() ([]model.OrderResponse, error)
	GetByID(id int) (model.OrderResponse, error)
	Update(name string, id int, itemReq []model.OrderItemRequest) error
	Delete(id int) error
	UpdateStatus(id int, status string) error
	NumberOfOrders(startDate, endDate interface{}) (model.NumberOfOrderedItemsResponse, error)
	GetPriceMap(names []string) (map[string]float64, error)
}

type Order struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *Order {
	return &Order{db: db}
}

var (
	ErrNotEnoughStock   = errors.New("insufficient_inventory")
	ErrMenuItemNotFound = errors.New("menu_item_not_found")
)

func (o *Order) Add(name string, itemReq []model.OrderItemRequest) (int, []model.InventoryUpdate, error) {
	tx, err := o.db.Begin()
	if err != nil {
		return 0, []model.InventoryUpdate{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ingredientNeeds := make(map[int]float64)
	totalAmount := 0.0

	for _, item := range itemReq {
		var price float64
		if err := tx.QueryRow(`
			SELECT price 
			FROM menu_items 
			WHERE name = $1
		`, item.MenuItemID).Scan(&price); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, []model.InventoryUpdate{}, fmt.Errorf("%w: menu item '%s' not found", ErrMenuItemNotFound, item.MenuItemID)
			}
			return 0, []model.InventoryUpdate{}, err
		}
		totalAmount += price * float64(item.Quantity)

		rows, err := tx.Query(`
			SELECT inventory_id, quantity
			FROM menu_item_ingredients
			JOIN menu_items ON menu_item_ingredients.menu_item_id = menu_items.menu_item_id
			WHERE menu_items.name = $1
		`, item.MenuItemID)
		if err != nil {
			return 0, []model.InventoryUpdate{}, err
		}

		for rows.Next() {
			var inventoryID int
			var quantityPerPortion float64
			if err := rows.Scan(&inventoryID, &quantityPerPortion); err != nil {
				return 0, []model.InventoryUpdate{}, err
			}
			ingredientNeeds[inventoryID] += quantityPerPortion * float64(item.Quantity)
		}
		rows.Close()
	}

	for inventoryID, neededQty := range ingredientNeeds {
		var currentStock float64
		if err := tx.QueryRow(`
			SELECT stock_level 
			FROM inventory 
			WHERE inventory_id = $1
		`, inventoryID).Scan(&currentStock); err != nil {
			return 0, []model.InventoryUpdate{}, err
		}
		if currentStock < neededQty {
			return 0, []model.InventoryUpdate{}, fmt.Errorf("%w: not enough stock for ingredient %d: need %.2f, have %.2f", ErrNotEnoughStock, inventoryID, neededQty, currentStock)
		}
	}

	for inventoryID, usedQty := range ingredientNeeds {
		_, err := tx.Exec(`
			UPDATE inventory 
			SET stock_level = stock_level - $1 
			WHERE inventory_id = $2
		`, usedQty, inventoryID)
		if err != nil {
			return 0, []model.InventoryUpdate{}, err
		}
	}

	var orderID int
	err = tx.QueryRow(`
				INSERT INTO orders(customer_name, status, total_amount)
				VALUES($1, 'active', $2)
				RETURNING order_id
			`, name, totalAmount).Scan(&orderID)
	if err != nil {
		return 0, []model.InventoryUpdate{}, err
	}

	for _, item := range itemReq {
		var price float64
		err := tx.QueryRow(`
			SELECT price 
			FROM menu_items 
			WHERE name = $1
		`, item.MenuItemID).Scan(&price)
		if err != nil {
			return 0, []model.InventoryUpdate{}, err
		}

		var menuItemId int
		err = tx.QueryRow(`
			SELECT menu_item_id 
			FROM menu_items 
			WHERE name = $1
		`, item.MenuItemID).Scan(&menuItemId)
		if err != nil {
			return 0, []model.InventoryUpdate{}, err
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (menu_item_id, order_id, customizations, price_at_order_time, quantity)
			VALUES($1, $2, '{}', $3, $4)
		`, menuItemId, orderID, price, item.Quantity)
		if err != nil {
			return 0, []model.InventoryUpdate{}, err
		}
	}

	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status)
		VALUES($1, 'active')
	`, orderID)
	if err != nil {
		return 0, []model.InventoryUpdate{}, err
	}

	var updates []model.InventoryUpdate
	for inventoryID, usedQty := range ingredientNeeds {
		_, err := tx.Exec(`UPDATE inventory SET stock_level = stock_level - $1 WHERE inventory_id = $2`, usedQty, inventoryID)
		if err != nil {
			return 0, nil, err
		}

		var name string
		var remaining float64
		if err := tx.QueryRow(`SELECT name, stock_level FROM inventory WHERE inventory_id = $1`, inventoryID).Scan(&name, &remaining); err != nil {
			return 0, nil, err
		}

		updates = append(updates, model.InventoryUpdate{
			IngredientID: inventoryID,
			Name:         name,
			QuantityUsed: usedQty,
			Remaining:    remaining,
		})
	}

	return orderID, updates, nil
}

func (o *Order) GetAll() ([]model.OrderResponse, error) {
	query := `
			SELECT 
		o.order_id,
		o.customer_name,
		o.status,
		o.order_date AS created_at,
		json_agg(json_build_object(
			'product_id', mi.name,
			'quantity', oi.quantity
		) ORDER BY oi.order_item_id) AS items
		FROM orders o
		JOIN order_items oi ON o.order_id = oi.order_id
		JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
		GROUP BY o.order_id, o.customer_name, o.status, o.order_date
		ORDER BY o.order_id;
	`

	rows, err := o.db.Query(query)
	if err != nil {
		return []model.OrderResponse{}, err
	}
	defer rows.Close()

	var orders []model.OrderResponse
	for rows.Next() {
		var order model.OrderResponse
		var itemsRow []byte

		if err := rows.Scan(&order.OrderID, &order.CustomerName, &order.Status, &order.CreatedAt, &itemsRow); err != nil {
			return []model.OrderResponse{}, err
		}

		loc, _ := time.LoadLocation("Asia/Almaty")
		order.CreatedAt = order.CreatedAt.In(loc)

		if err := json.Unmarshal(itemsRow, &order.Items); err != nil {
			return []model.OrderResponse{}, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (o *Order) GetByID(id int) (model.OrderResponse, error) {
	query := `
		SELECT 
		o.order_id,
		o.customer_name,
		o.status,
		o.order_date AS created_at,
		json_agg(json_build_object(
			'product_id', mi.name,
			'quantity', oi.quantity
		) ORDER BY oi.order_item_id) AS items
		FROM orders o
		JOIN order_items oi ON o.order_id = oi.order_id
		JOIN menu_items mi ON oi.menu_item_id = mi.menu_item_id
		WHERE o.order_id = $1
		GROUP BY o.order_id, o.customer_name, o.status, o.order_date
		ORDER BY o.order_id;
	`

	rows, err := o.db.Query(query, id)
	if err != nil {
		return model.OrderResponse{}, err
	}
	defer rows.Close()

	var order model.OrderResponse
	var found bool

	for rows.Next() {
		found = true
		var itemsRow []byte
		if err := rows.Scan(&order.OrderID, &order.CustomerName, &order.Status, &order.CreatedAt, &itemsRow); err != nil {
			return model.OrderResponse{}, err
		}

		loc, _ := time.LoadLocation("Asia/Almaty")
		order.CreatedAt = order.CreatedAt.In(loc)

		if err := json.Unmarshal(itemsRow, &order.Items); err != nil {
			return model.OrderResponse{}, err
		}
	}

	if !found {
		return model.OrderResponse{}, fmt.Errorf("order with id %d not found", id)
	}

	return order, nil
}

func (o *Order) Update(name string, id int, itemReq []model.OrderItemRequest) error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	ingredientNeeds := make(map[int]float64)
	totalAmount := 0.0

	for _, item := range itemReq {
		var price float64
		if err := tx.QueryRow(`SELECT price FROM menu_items WHERE name = $1`, item.MenuItemID).Scan(&price); err != nil {
			return err
		}
		totalAmount += price * float64(item.Quantity)

		rows, err := tx.Query(`
			SELECT inventory_id, quantity
			FROM menu_item_ingredients
			JOIN menu_items ON menu_item_ingredients.menu_item_id = menu_items.menu_item_id
			WHERE menu_items.name = $1
		`, item.MenuItemID)
		if err != nil {
			return err
		}

		for rows.Next() {
			var inventoryID int
			var quantityPerPortion float64
			if err := rows.Scan(&inventoryID, &quantityPerPortion); err != nil {
				return err
			}
			ingredientNeeds[inventoryID] += quantityPerPortion * float64(item.Quantity)
		}
		rows.Close()
	}

	for inventoryID, neededQty := range ingredientNeeds {
		var currentStock float64
		if err := tx.QueryRow(`SELECT stock_level FROM inventory WHERE inventory_id = $1`, inventoryID).Scan(&currentStock); err != nil {
			return err
		}
		if currentStock < neededQty {
			return fmt.Errorf("not enough stock for ingredient %d: need %.2f, have %.2f", inventoryID, neededQty, currentStock)
		}
	}

	for inventoryID, usedQty := range ingredientNeeds {
		_, err := tx.Exec(`UPDATE inventory SET stock_level = stock_level - $1 WHERE inventory_id = $2`, usedQty, inventoryID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		UPDATE orders
		SET customer_name = $1, status = 'active', total_amount = $2, order_date = NOW()
		WHERE order_id = $3
	`, name, totalAmount, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		DELETE FROM order_items WHERE order_id = $1
	`, id)
	if err != nil {
		return err
	}

	for _, item := range itemReq {
		var price float64
		err := tx.QueryRow(`
			SELECT price
			FROM menu_items
			WHERE name = $1
		`, item.MenuItemID).Scan(&price)
		if err != nil {
			return err
		}

		var menuItemId int
		err = tx.QueryRow(`
			SELECT menu_item_id
			FROM menu_items
			WHERE name = $1
		`, item.MenuItemID).Scan(&menuItemId)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (menu_item_id, order_id, customizations, price_at_order_time, quantity)
			VALUES($1, $2, '{}', $3, $4)
		`, menuItemId, id, price, item.Quantity)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status)
		VALUES($1, 'active')
	`, id)
	if err != nil {
		return err
	}

	return nil
}

func (o *Order) Delete(id int) error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `DELETE FROM orders WHERE order_id = $1`
	if _, err = tx.Exec(query, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (o *Order) NumberOfOrders(startDate, endDate interface{}) (model.NumberOfOrderedItemsResponse, error) {
	query := `
		SELECT
			mi.name,
			COALESCE(SUM(oi.quantity), 0) AS total_quantity
		FROM menu_items mi
		LEFT JOIN order_items oi ON mi.menu_item_id = oi.menu_item_id
		LEFT JOIN orders o ON oi.order_id = o.order_id
		WHERE
			($1::DATE IS NULL OR o.order_date >= $1::DATE)
			AND ($2::DATE IS NULL OR o.order_date <= $2::DATE)
		GROUP BY mi.name
		ORDER BY mi.name;
	`

	rows, err := o.db.Query(query, startDate, endDate)
	if err != nil {
		return model.NumberOfOrderedItemsResponse{}, err
	}

	orderCount := make(model.NumberOfOrderedItemsResponse)

	for rows.Next() {
		var name string
		var quantity int

		if err := rows.Scan(&name, &quantity); err != nil {
			return model.NumberOfOrderedItemsResponse{}, err
		}
		orderCount[name] = quantity
	}
	return orderCount, nil
}

func (o *Order) UpdateStatus(id int, status string) error {
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		UPDATE order_status_history
		SET status = $2, changed_at = $3
		WHERE order_id = $1
	`

	if _, err = tx.Exec(query, id, status, time.Now()); err != nil {
		return err
	}

	query = `
		UPDATE orders
		SET status = $2
		WHERE order_id = $1
	`

	if _, err = tx.Exec(query, id, status); err != nil {
		return err
	}

	return tx.Commit()
}

func (o *Order) GetPriceMap(names []string) (map[string]float64, error) {
	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names))

	for i, name := range names {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	query := fmt.Sprintf(`
		SELECT name, price
		FROM menu_items
		WHERE name IN (%s)
	`, strings.Join(placeholders, ","))

	rows, err := o.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	priceMap := make(map[string]float64)
	for rows.Next() {
		var name string
		var price float64
		if err := rows.Scan(&name, &price); err != nil {
			return nil, err
		}
		priceMap[name] = price
	}

	return priceMap, nil
}
