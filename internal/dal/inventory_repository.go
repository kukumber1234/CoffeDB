package dal

import (
	"database/sql"
	"errors"
	"time"

	model "frappuccino/models"
)

type InventoryRepository interface {
	Add(inventoryItem model.InventoryItem) (model.InventoryItem, error)
	GetAll() ([]model.InventoryItem, error)
	GetByID(id int) (model.InventoryItem, error)
	Update(inventoryItem model.InventoryItem) error
	Delete(id int) error
	CountInventory(sortBy string, page, pageSize int) (model.CountInventory, error)
	AddTransaction(inventoryID int, quantity float64) error
}

type Inventory struct {
	db *sql.DB
}

func NewInventoryRepo(db *sql.DB) *Inventory {
	return &Inventory{db: db}
}

func (i *Inventory) Add(inventoryItem model.InventoryItem) (model.InventoryItem, error) {
	tx, err := i.db.Begin()
	if err != nil {
		return model.InventoryItem{}, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO inventory (name, stock_level, reorder_level)
		VALUES ($1, $2, $3) RETURNING inventory_id, last_updated
	`

	var inventoryItemID int
	var lastUpdated time.Time
	if err = tx.QueryRow(query, inventoryItem.Name, *inventoryItem.StockLevel, inventoryItem.ReorderLevel).Scan(&inventoryItemID, &lastUpdated); err != nil {
		return model.InventoryItem{}, err
	}

	if err = tx.Commit(); err != nil {
		return model.InventoryItem{}, err
	}

	inventoryItem.IngredientID = &inventoryItemID
	inventoryItem.LastUpdated = lastUpdated

	if err := i.AddTransaction(inventoryItemID, *inventoryItem.StockLevel); err != nil {
		return model.InventoryItem{}, err
	}
	return inventoryItem, nil
}

func (i *Inventory) GetAll() ([]model.InventoryItem, error) {
	query := `
		SELECT inventory_id, name, stock_level, reorder_level, last_updated
		FROM inventory
	`

	rows, err := i.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventoryItems []model.InventoryItem
	for rows.Next() {
		inventoryItem := model.InventoryItem{}
		var id int
		var stock float64
		var lastUpdated time.Time

		if err := rows.Scan(&id, &inventoryItem.Name, &stock, &inventoryItem.ReorderLevel, &lastUpdated); err != nil {
			return nil, err
		}

		loc, _ := time.LoadLocation("Asia/Almaty")
		lastUpdated = lastUpdated.In(loc)

		inventoryItem.IngredientID = &id
		inventoryItem.StockLevel = &stock
		inventoryItem.LastUpdated = lastUpdated

		inventoryItems = append(inventoryItems, inventoryItem)
	}
	return inventoryItems, nil
}

func (i *Inventory) GetByID(id int) (model.InventoryItem, error) {
	query := `
		SELECT inventory_id, name, stock_level, reorder_level, last_updated
		FROM inventory
		WHERE inventory_id = $1
	`
	row := i.db.QueryRow(query, id)

	inventoryItem := model.InventoryItem{}
	var invID int
	var stock float64
	var lastUpdated time.Time

	err := row.Scan(&invID, &inventoryItem.Name, &stock, &inventoryItem.ReorderLevel, &lastUpdated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.InventoryItem{}, errors.New("inventory item not found")
		}
		return model.InventoryItem{}, err
	}

	loc, _ := time.LoadLocation("Asia/Almaty")
	lastUpdated = lastUpdated.In(loc)

	inventoryItem.IngredientID = &invID
	inventoryItem.StockLevel = &stock
	inventoryItem.LastUpdated = lastUpdated

	return inventoryItem, nil
}

func (i *Inventory) Update(inventoryItem model.InventoryItem) error {
	tx, err := i.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ingredient, err := i.GetByID(*inventoryItem.IngredientID)
	if inventoryItem.StockLevel == nil {
		inventoryItem.StockLevel = ingredient.StockLevel
	}
	if inventoryItem.Name == "" {
		inventoryItem.Name = ingredient.Name
	}
	if inventoryItem.ReorderLevel == nil {
		inventoryItem.ReorderLevel = ingredient.ReorderLevel
	}

	query := `
		UPDATE inventory
		SET name = $1, stock_level = $2, reorder_level = $3
		WHERE inventory_id = $4
	`

	if _, err = tx.Exec(query, inventoryItem.Name, inventoryItem.StockLevel, inventoryItem.ReorderLevel, inventoryItem.IngredientID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	if inventoryItem.StockLevel != nil {
		if err := i.AddTransaction(*inventoryItem.IngredientID, *inventoryItem.StockLevel); err != nil {
			return err
		}
	}
	return nil
}

func (i *Inventory) Delete(id int) error {
	tx, err := i.db.Begin()
	if err != nil {
		return err
	}

	query := `DELETE FROM inventory WHERE inventory_id = $1`
	if _, err = tx.Exec(query, id); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (i *Inventory) CountInventory(sortBy string, page, pageSize int) (model.CountInventory, error) {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	if sortBy != "name" && sortBy != "quantity" {
		sortBy = "name"
	}

	var totalItem int
	err := i.db.QueryRow(`SELECT COUNT(*) FROM inventory`).Scan(&totalItem)
	if err != nil {
		return model.CountInventory{}, err
	}

	query := `
		SELECT name, stock_level
		FROM inventory
		ORDER BY 
			CASE WHEN $1 = 'quantity' THEN stock_level ELSE NULL END DESC,
			CASE WHEN $1 = 'name' THEN name ELSE NULL END
		LIMIT $2 OFFSET $3
	`

	rows, err := i.db.Query(query, sortBy, pageSize, offset)
	if err != nil {
		return model.CountInventory{}, err
	}
	defer rows.Close()

	var items []model.Data

	for rows.Next() {
		var item model.Data
		var stockLevel float64

		if err := rows.Scan(&item.Name, &stockLevel); err != nil {
			return model.CountInventory{}, err
		}
		item.Quantity = int(stockLevel)
		items = append(items, item)
	}

	totalPage := (totalItem + pageSize - 1) / pageSize
	hasNextPage := page < totalPage

	response := model.CountInventory{
		CurrentPage: page,
		HasNextPage: hasNextPage,
		PageSize:    pageSize,
		TotalPages:  totalPage,
		Data:        items,
	}

	return response, nil
}

func (i *Inventory) AddTransaction(inventoryID int, quantity float64) error {
	tx, err := i.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	query := `
		INSERT INTO inventory_transactions (inventory_id, quantity)
		VALUES ($1, $2) RETURNING transaction_id, transaction_date
	`

	var transactionID int
	var transactionDate time.Time

	if err = tx.QueryRow(query, inventoryID, quantity).Scan(&transactionID, &transactionDate); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
