package handler

import (
	"encoding/json"
	"frappuccino/config"
	"frappuccino/internal/service"
	"frappuccino/models"
	"net/http"
	"strconv"
)

type OrderHandler struct {
	OrderService service.OrdersService
}

func NewOrderHandler(service service.OrdersService) *OrderHandler {
	return &OrderHandler{OrderService: service}
}

func (o *OrderHandler) Add(w http.ResponseWriter, r *http.Request) {
	var orderRequest models.OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		SendResponse("Failed to decode order", err, http.StatusInternalServerError, w)
		return
	}

	orderID, err := o.OrderService.Add(orderRequest.CustomerName, orderRequest.Orders)
	if err != nil {
		SendResponse("Failed to add order", err, http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Order placed successfully",
		"order_id": orderID,
	})
}

func (o *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	items, err := o.OrderService.GetAll()
	if err != nil {
		SendResponse("Failed to load orders", err, http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(items); err != nil {
		return
	}
}

func (o *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Error convert string to int", err, http.StatusNotFound, w)
		return
	}
	item, err := o.OrderService.GetByID(id)
	if err != nil {
		SendResponse("Order item not found", err, http.StatusNotFound, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		return
	}
}

func (o *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Update")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}
	if err = o.OrderService.CloseOrder(id); err != nil {
		SendResponse("Failed to close order", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully updated order", nil, http.StatusOK, w)
}

func (o *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Delete")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}

	if err := o.OrderService.Delete(id); err != nil {
		SendResponse("Failed to delete item", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully deleted order", nil, http.StatusOK, w)
}

func (o *OrderHandler) NumberOfOrders(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	startDate := query.Get("startDate")
	endDate := query.Get("endDate")

	item, err := o.OrderService.NumberOfOrders(StringOrNil(startDate), StringOrNil(endDate))
	if err != nil {
		SendResponse("Failed to count order", err, http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		return
	}
}

func (o *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}
	var orderRequest models.OrderRequest

	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		SendResponse("Failed to decode order", err, http.StatusInternalServerError, w)
		return
	}

	if err := o.OrderService.Update(orderRequest.CustomerName, id, orderRequest.Orders); err != nil {
		SendResponse("Failed to update order", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully updated order", nil, http.StatusOK, w)
}

func StringOrNil(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
