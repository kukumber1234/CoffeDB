package service

import (
	"errors"
	"frappuccino/internal/dal"
	model "frappuccino/models"
)

type OrdersService interface {
	Add(name string, itemReq []model.OrderItemRequest) (int, error)
	GetAll() ([]model.OrderResponse, error)
	GetByID(id int) (model.OrderResponse, error)
	// Update(order model.Order) error
	CloseOrder(id int) error
	Delete(id int) error
	NumberOfOrders(startDate, endDate interface{}) (model.NumberOfOrderedItemsResponse, error)
}

type Order struct {
	repository dal.OrderRepository
}

func NewOrderService(dataAccess dal.OrderRepository) *Order {
	return &Order{repository: dataAccess}
}

func (o *Order) Add(name string, itemReq []model.OrderItemRequest) (int, error) {
	return o.repository.Add(name, itemReq)
}

func (o *Order) GetAll() ([]model.OrderResponse, error) {
	return o.repository.GetAll()
}

func (o *Order) GetByID(id int) (model.OrderResponse, error) {
	return o.repository.GetByID(id)
}

// func (o *Order) Update(order model.Order) error {
// 	return o.repository.Update(&order)
// }

func (o *Order) CloseOrder(id int) error {
	return o.repository.UpdateStatus(id, "closed")
}

func (o *Order) Delete(id int) error {
	order, err := o.repository.GetByID(id)
	if err != nil {
		return err
	}
	if order.OrderID != id {
		return errors.New("order ID not match")
	}
	return o.repository.Delete(id)
}

func (o *Order) NumberOfOrders(startDate, endDate interface{}) (model.NumberOfOrderedItemsResponse, error) {
	return o.repository.NumberOfOrders(startDate, endDate)
}
