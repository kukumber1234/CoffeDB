package service

import (
	"errors"
	"frappuccino/internal/dal"
	"frappuccino/models"
)

type InventoryService interface {
	Add(inventoryItem models.InventoryItem) error
	Get() ([]models.InventoryItem, error)
	GetByID(id int) (models.InventoryItem, error)
	Update(id int, inventoryItem models.InventoryItem) error
	Delete(id int) error
	CountInventory(sortBy string, page, pageSize int) (models.CountInventory, error)
}

type Inventory struct {
	repository dal.InventoryRepository
}

func NewInventoryService(inventory dal.InventoryRepository) *Inventory {
	return &Inventory{repository: inventory}
}

func (s *Inventory) Add(inventoryItem models.InventoryItem) error {
	if inventoryItem.Name == "" {
		return errors.New("ingredient name can not be empty")
	}

	if inventoryItem.StockLevel == nil {
		return errors.New("ingredient stock level can not be empty")
	}

	if inventoryItem.ReorderLevel == nil {
		return errors.New("ingredient reorder level can not be empty")
	}

	if *inventoryItem.StockLevel <= 0 {
		return errors.New("ingredient quantity can not be lower or equal than 0")
	}

	if *inventoryItem.ReorderLevel <= 0 {
		return errors.New("ingredient quantity can not be lower or equal than 0")
	}
	if _, err := s.repository.Add(inventoryItem); err != nil {
		return err
	}
	return nil
}

func (s *Inventory) Get() ([]models.InventoryItem, error) {
	return s.repository.GetAll()
}

func (s *Inventory) GetByID(id int) (models.InventoryItem, error) {
	return s.repository.GetByID(id)
}

func (s *Inventory) Update(id int, inventoryItem models.InventoryItem) error {
	if id <= 0 {
		return errors.New("id can not be empty and less or equal to zero")
	}
	if _, err := s.repository.GetByID(id); err != nil {
		return err
	}

	if inventoryItem.IngredientID != nil {
		return errors.New("ingredient id can not be changed")
	}
	inventoryItem.IngredientID = &id

	if inventoryItem.Name == "" {
		return errors.New("ingredient name can not be empty")
	}

	if inventoryItem.StockLevel != nil && *inventoryItem.ReorderLevel <= 0 {
		return errors.New("ingredient quantity can not be lower or equal than 0")
	}

	if inventoryItem.StockLevel != nil && *inventoryItem.StockLevel <= 0 {
		return errors.New("ingredient quantity can not be lower or equal than 0")
	}
	return s.repository.Update(inventoryItem)
}

func (s *Inventory) Delete(id int) error {
	if id <= 0 {
		return errors.New("id can not be empty and less or equal to zero")
	}
	if _, err := s.repository.GetByID(id); err != nil {
		return err
	}

	return s.repository.Delete(id)
}

func (s *Inventory) CountInventory(sortBy string, page, pageSize int) (models.CountInventory, error) {
	return s.repository.CountInventory(sortBy, page, pageSize)
}
