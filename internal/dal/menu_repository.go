package dal

import (
	"database/sql"
	"fmt"
	"time"

	model "frappuccino/models"

	"github.com/lib/pq"
)

type MenuRepository interface {
	GetAll() ([]model.MenuRequest, error)
	GetByID(id int) (model.MenuRequest, error)
	Save(item model.MenuItem, menuIngredients []model.MenuInventory) error
	Update(item model.MenuItem, menuIngredients []model.MenuInventory) error
	Delete(id int) error
}

type Menu struct {
	db *sql.DB
}

func NewMenuRepo(db *sql.DB) *Menu {
	return &Menu{db: db}
}

func (f *Menu) GetAll() ([]model.MenuRequest, error) {
	query := `
		SELECT
			menu_items.menu_item_id,
			menu_items.name, 
			menu_items.description, 
			menu_item_ingredients.quantity,
			menu_items.price,
			menu_items.tags,
			inventory.name
		FROM 
			menu_items
		JOIN 
			menu_item_ingredients ON menu_items.menu_item_id = menu_item_ingredients.menu_item_id
		JOIN 
			inventory ON menu_item_ingredients.inventory_id = inventory.inventory_id
		`

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menuMap := make(map[string]*model.MenuRequest)
	for rows.Next() {
		var id int
		var menuName, description, inventoryName string
		var price, quantity float64
		var tagsString []string

		if err := rows.Scan(&id, &menuName, &description, &quantity, &price, pq.Array(&tagsString), &inventoryName); err != nil {
			return nil, err
		}

		if _, exists := menuMap[menuName]; !exists {
			menuMap[menuName] = &model.MenuRequest{
				Menu: model.MenuItem{
					ID:          id,
					Name:        menuName,
					Description: description,
					Price:       price,
					Tags:        tagsString,
				},
			}
		}

		menuMap[menuName].MenuIngredients = append(menuMap[menuName].MenuIngredients, model.MenuInventory{
			Inventory: model.InventoryMenuRequest{
				Name: inventoryName,
			},
			Quantity: quantity,
		})
	}

	var menuReq []model.MenuRequest
	for _, value := range menuMap {
		menuReq = append(menuReq, *value)
	}

	return menuReq, nil
}

func (f *Menu) GetByID(id int) (model.MenuRequest, error) {
	query := `
		SELECT
			menu_items.menu_item_id,
			menu_items.name, 
			menu_items.description, 
			menu_item_ingredients.quantity,
			menu_items.price,
			menu_items.tags,
			inventory.name
		FROM 
			menu_items
		JOIN 
			menu_item_ingredients ON menu_items.menu_item_id = menu_item_ingredients.menu_item_id
		JOIN 
			inventory ON menu_item_ingredients.inventory_id = inventory.inventory_id
		WHERE
			menu_items.menu_item_id = $1
		`
	rows, err := f.db.Query(query, id)
	if err != nil {
		return model.MenuRequest{}, nil
	}
	defer rows.Close()

	menuItem := model.MenuRequest{}
	menuItem.MenuIngredients = []model.MenuInventory{}

	for rows.Next() {
		var quantity float64
		var tags []string
		var ingredientName string

		err := rows.Scan(
			&menuItem.Menu.ID,
			&menuItem.Menu.Name,
			&menuItem.Menu.Description,
			&quantity,
			&menuItem.Menu.Price,
			pq.Array(&tags),
			&ingredientName,
		)
		if err != nil {
			return model.MenuRequest{}, err
		}

		menuItem.Menu.Tags = tags
		menuItem.MenuIngredients = append(menuItem.MenuIngredients, model.MenuInventory{
			Inventory: model.InventoryMenuRequest{
				Name: ingredientName,
			},
			Quantity: quantity,
		})
	}

	if len(menuItem.MenuIngredients) == 0 {
		return model.MenuRequest{}, sql.ErrNoRows
	}

	return menuItem, nil
}

func (f *Menu) Save(item model.MenuItem, menuIngredients []model.MenuInventory) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}

	menuQuery := `INSERT INTO menu_items (name, description, price, tags)
				  VALUES ($1, $2, $3, $4) RETURNING menu_item_id`
	var menuItemID int
	err = tx.QueryRow(menuQuery, item.Name, item.Description, item.Price, pq.Array(item.Tags)).Scan(&menuItemID)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, menuIngredient := range menuIngredients {

		var ingredientID int
		ingredientIDQuery := `SELECT inventory_id FROM inventory WHERE name = $1`
		err = tx.QueryRow(ingredientIDQuery, menuIngredient.Inventory.Name).Scan(&ingredientID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("no such item in inventory: '%s'", menuIngredient.Inventory.Name)
		}

		menuItemQuery := `INSERT INTO menu_item_ingredients (inventory_id, menu_item_id, quantity)
						  VALUES ($1, $2, $3)`
		_, err = tx.Exec(menuItemQuery, ingredientID, menuItemID, menuIngredient.Quantity)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	priceHistoryQuery := `INSERT INTO price_history (menu_item_id, new_price, changed_at)
						  VALUES($1, $2, $3)`

	_, err = tx.Exec(priceHistoryQuery, menuItemID, item.Price, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (f *Menu) Update(item model.MenuItem, menuIngredients []model.MenuInventory) error {
	tx, err := f.db.Begin()
	if err != nil {
		fmt.Println("HERE 1")
		return err
	}

	var oldPrice float64

	oldPriceQuery := `SELECT price FROM menu_items
				 	  WHERE menu_item_id = $1`

	err = tx.QueryRow(oldPriceQuery, item.ID).Scan(&oldPrice)
	if err != nil {
		fmt.Println("HERE 2")
		return err
	}

	if oldPrice != item.Price {
		updatePriceQuery := `UPDATE price_history
						 SET old_price = $1, new_price = $2, changed_at = $3
						 WHERE menu_item_id = $4`

		_, err = tx.Exec(updatePriceQuery, oldPrice, item.Price, time.Now(), item.ID)
		if err != nil {
			fmt.Println("HERE 3")
			tx.Rollback()
			return err
		}
	}

	query := `
		UPDATE menu_items
		SET name = $1, description = $2, price = $3, tags = $4
		WHERE menu_item_id = $5
	`

	tags := pq.Array(item.Tags)

	_, err = tx.Exec(query,
		item.Name,
		item.Description,
		item.Price,
		tags,
		item.ID,
	)
	if err != nil {
		fmt.Println("HERE 4")
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`, item.ID)
	if err != nil {
		fmt.Println("HERE 5")
		tx.Rollback()
		return err
	}

	for _, ingredient := range menuIngredients {
		var inventoryID int

		queryInventoryID := `SELECT inventory_id FROM inventory WHERE name = $1`
		err := tx.QueryRow(queryInventoryID, ingredient.Inventory.Name).Scan(&inventoryID)
		if err != nil {
			fmt.Println("error: ", err)
			tx.Rollback()
			return err
		}

		query := `
			INSERT INTO menu_item_ingredients(inventory_id, menu_item_id, quantity)
			VALUES($1, $2, $3)
		`
		_, err = tx.Exec(query, inventoryID, item.ID, ingredient.Quantity)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (f *Menu) Delete(id int) error {
	tx, err := f.db.Begin()
	if err != nil {
		return err
	}

	query := `DELETE FROM menu_items WHERE menu_item_id = $1`
	if _, err = tx.Exec(query, id); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
