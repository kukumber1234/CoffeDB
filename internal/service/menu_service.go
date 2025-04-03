package service

import (
	// "fmt"
	// "frapo/config"

	"errors"
	"fmt"

	"frappuccino/config"
	dal "frappuccino/internal/dal"
	model "frappuccino/models"
)

type MenuService interface {
	Add(item model.MenuItem, menuIngredients []model.MenuInventory) error
	Get() ([]model.MenuRequest, error)
	GetByID(id int) (*model.MenuRequest, error)
	Update(item model.MenuItem, menuIngredients []model.MenuInventory) error
	Delete(id int) error
}

type Menu struct {
	dataAccess dal.MenuRepository
}

func NewFileMenuService(dataAccess dal.MenuRepository) *Menu {
	return &Menu{dataAccess: dataAccess}
}

func (f *Menu) Add(item model.MenuItem, menuIngredients []model.MenuInventory) error {
	return f.dataAccess.Save(item, menuIngredients)
}

func (f *Menu) Get() ([]model.MenuRequest, error) {
	return f.dataAccess.GetAll()
}

func (f *Menu) GetByID(id int) (*model.MenuRequest, error) {
	items, err := f.dataAccess.GetByID(id)
	if err != nil {
		config.Logger.Info("menu item not found")
		return nil, fmt.Errorf("menu item not found")
	}
	return &items, nil
}

func (f *Menu) Update(item model.MenuItem, menuIngredients []model.MenuInventory) error {
	if item.ID <= 0 {
		return errors.New("id can not be empty or zero")
	}

	if item.Name == "" {
		return errors.New("name can not be empty")
	}

	if item.Price < 0 {
		return errors.New("price can not bo lower than 0")
	}

	if len(item.Tags) == 0 {
		return errors.New("tags can not be empty")
	}

	if item.Description == "" {
		return errors.New("description can not be equal")
	}

	for _, val := range menuIngredients {
		if val.Quantity <= 0 {
			return errors.New("quantity can not be equal or less than 0")
		}

		if val.Inventory.Name == "" {
			return errors.New("inventory name can not be empty")
		}
	}

	return f.dataAccess.Update(item, menuIngredients)
}

func (f *Menu) Delete(id int) error {
	if id <= 0 {
		return errors.New("id can not be empty or zero")
	}

	return f.dataAccess.Delete(id)
}
