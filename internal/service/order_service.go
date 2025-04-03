package service

import (
	"errors"

	"frappuccino/internal/dal"
	model "frappuccino/models"
)

type OrdersService interface {
	Add(name string, itemReq []model.OrderItemRequest) (int, []model.InventoryUpdate, error)
	GetAll() ([]model.OrderResponse, error)
	GetByID(id int) (model.OrderResponse, error)
	Update(name string, id int, itemReq []model.OrderItemRequest) error
	CloseOrder(id int) error
	Delete(id int) error
	NumberOfOrders(startDate, endDate interface{}) (model.NumberOfOrderedItemsResponse, error)
	BatchProcessOrders(request model.BatchOrderRequest) (model.BatchOrderResponse, error)
}

type Order struct {
	repository dal.OrderRepository
}

func NewOrderService(dataAccess dal.OrderRepository) *Order {
	return &Order{repository: dataAccess}
}

func (o *Order) Add(name string, itemReq []model.OrderItemRequest) (int, []model.InventoryUpdate, error) {
	return o.repository.Add(name, itemReq)
}

func (o *Order) GetAll() ([]model.OrderResponse, error) {
	return o.repository.GetAll()
}

func (o *Order) GetByID(id int) (model.OrderResponse, error) {
	return o.repository.GetByID(id)
}

func (o *Order) Update(name string, id int, itemReq []model.OrderItemRequest) error {
	return o.repository.Update(name, id, itemReq)
}

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

func (s *Order) BatchProcessOrders(request model.BatchOrderRequest) (model.BatchOrderResponse, error) {
	var (
		processedOrders  []model.ProcessedOrder
		totalRevenue     float64
		accepted         int
		rejected         int
		inventoryUpdates []model.InventoryUpdate
	)

	var allItemNames []string
	for _, order := range request.Orders {
		for _, item := range order.Items {
			allItemNames = append(allItemNames, item.MenuItemName)
		}
	}

	priceMap, err := s.repository.GetPriceMap(allItemNames)
	if err != nil {
		return model.BatchOrderResponse{}, err
	}

	for _, order := range request.Orders {
		mappedItems := mapToStandardItemReq(order.Items)

		orderID, updates, err := s.Add(order.CustomerName, mappedItems)
		if err != nil {
			status := "rejected"
			reason := "unknown_error"

			if errors.Is(err, dal.ErrNotEnoughStock) {
				reason = "insufficient_inventory"
			} else if errors.Is(err, dal.ErrMenuItemNotFound) {
				reason = "menu_item_not_found"
			}

			processedOrders = append(processedOrders, model.ProcessedOrder{
				CustomerName: order.CustomerName,
				Status:       status,
				Reason:       reason,
			})
			rejected++
			continue
		}

		total := calculateTotal(order.Items, priceMap)
		totalRevenue += total
		accepted++

		processedOrders = append(processedOrders, model.ProcessedOrder{
			OrderID:      orderID,
			CustomerName: order.CustomerName,
			Status:       "accepted",
			Total:        total,
		})

		inventoryUpdates = append(inventoryUpdates, updates...)
	}

	return model.BatchOrderResponse{
		ProcessedOrders: processedOrders,
		Summary: model.BatchSummary{
			TotalOrders:      len(request.Orders),
			Accepted:         accepted,
			Rejected:         rejected,
			TotalRevenue:     totalRevenue,
			InventoryUpdates: inventoryUpdates, // можно позже заполнить
		},
	}, nil
}

func mapToStandardItemReq(batchItems []model.OrderItemRequestBatch) []model.OrderItemRequest {
	var result []model.OrderItemRequest
	for _, item := range batchItems {
		result = append(result, model.OrderItemRequest{
			MenuItemID: item.MenuItemName,
			Quantity:   item.Quantity,
		})
	}
	return result
}

func calculateTotal(items []model.OrderItemRequestBatch, priceMap map[string]float64) float64 {
	total := 0.0
	for _, item := range items {
		if price, ok := priceMap[item.MenuItemName]; ok {
			total += price * float64(item.Quantity)
		}
	}
	return total
}
