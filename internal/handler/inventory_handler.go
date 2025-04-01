package handler

import (
	"encoding/json"
	"frappuccino/config"
	"frappuccino/internal/service"
	"frappuccino/models"
	"io"
	"net/http"
	"strconv"
)

type Inventory struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *Inventory {
	return &Inventory{inventoryService: inventoryService}
}

func (h *Inventory) Add(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Add")
	var newInventoryItem models.InventoryItem

	if err := json.NewDecoder(r.Body).Decode(&newInventoryItem); err != nil {
		SendResponse("Failed to decode body", err, http.StatusBadRequest, w)
		return
	}

	if err := h.inventoryService.Add(newInventoryItem); err != nil {
		SendResponse("Failed to add a new item to the inventory", err, http.StatusBadRequest, w)
		return
	}

	SendResponse("Added new item to the inventory", nil, http.StatusCreated, w)
}

func (h *Inventory) Get(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Get")
	items, err := h.inventoryService.Get()
	if err != nil {
		SendResponse("Failed to Get inventory", err, http.StatusBadRequest, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(items); err != nil {
		SendResponse("Failed to encode inventory", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully retrieved inventory", nil, http.StatusOK, w)
}

func (h *Inventory) GetByID(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "GetByID")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}

	item, err := h.inventoryService.GetByID(id)
	if err != nil {
		SendResponse("Failed to get inventory", err, http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		SendResponse("Failed to encode inventory", err, http.StatusInternalServerError, w)
		return
	}
	config.Logger.Info("Successfully retrieved inventory", "Action", "GetByID")
}

func (h *Inventory) Update(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Update")
	var updatedItem models.InventoryItem
	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendResponse("Failed to decode body", err, http.StatusBadRequest, w)
		return
	}
	if err := json.Unmarshal(body, &updatedItem); err != nil {
		SendResponse("Failed to decode body", err, http.StatusBadRequest, w)
		return
	}
	id, err := strconv.Atoi(r.PathValue("{id}"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}
	if err = h.inventoryService.Update(id, updatedItem); err != nil {
		SendResponse("Failed to update item", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully updated inventory", nil, http.StatusOK, w)
}

func (h *Inventory) Delete(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "Delete")
	id, err := strconv.Atoi(r.PathValue("{id}"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}

	if err := h.inventoryService.Delete(id); err != nil {
		SendResponse("Failed to delete item", err, http.StatusInternalServerError, w)
		return
	}
	SendResponse("Successfully deleted inventory", nil, http.StatusOK, w)
}

func (h *Inventory) CountInventory(w http.ResponseWriter, r *http.Request) {
	config.Logger.Info("Incoming Request Received", "Action", "CountInventory")
	query := r.URL.Query()

	sortBy := query.Get("sortBy")
	pageStr := query.Get("page")
	pageSizeStr := query.Get("pageSize")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	items, err := h.inventoryService.CountInventory(sortBy, page, pageSize)
	if err != nil {
		SendResponse("Failed to count inventory", err, http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(items); err != nil {
		SendResponse("Failed to encode inventory", err, http.StatusInternalServerError, w)
		return
	}
}
