package handler

import (
	"encoding/json"
	"frappuccino/models"
	"net/http"
	"strconv"

	"frappuccino/internal/service"
)

type MenuHandler struct {
	service service.MenuService
}

func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{service: service}
}

func (m *MenuHandler) Add(w http.ResponseWriter, r *http.Request) {
	var request models.MenuRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		SendResponse("Invalid request payload", err, http.StatusBadRequest, w)
		return
	}

	if err := m.service.Add(request.Menu, request.MenuIngredients); err != nil {
		SendResponse("Failed to add menu item", err, http.StatusInternalServerError, w)
		return
	}

	SendResponse("Menu item added successfully", nil, http.StatusCreated, w)
}

func (m *MenuHandler) Get(w http.ResponseWriter, r *http.Request) {
	items, err := m.service.Get()
	if err != nil {
		SendResponse("Failed to load menu", err, http.StatusInternalServerError, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(items); err != nil {
		return
	}
}

func (m *MenuHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Error convert string to int", err, http.StatusNotFound, w)
		return
	}
	item, err := m.service.GetByID(id)
	if err != nil {
		SendResponse("MenuHandler item not found", err, http.StatusNotFound, w)
		return
	}
	w.Header().Set("Content-type", "application/json")
	if err = json.NewEncoder(w).Encode(item); err != nil {
		return
	}
}

func (m *MenuHandler) Update(w http.ResponseWriter, r *http.Request) {
	var menuReq models.MenuRequest

	if err := json.NewDecoder(r.Body).Decode(&menuReq); err != nil {
		SendResponse("Invalid request payload", err, http.StatusBadRequest, w)
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}

	menuReq.Menu.ID = id

	if err := m.service.Update(menuReq.Menu, menuReq.MenuIngredients); err != nil {
		SendResponse("Failed to add menu item", err, http.StatusInternalServerError, w)
		return
	}

	w.WriteHeader(http.StatusOK)
	SendResponse("Menu updated successfully", nil, http.StatusCreated, w)
}

func (m *MenuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		SendResponse("Failed to convert id to int", err, http.StatusInternalServerError, w)
		return
	}
	err = m.service.Delete(id)
	if err != nil {
		SendResponse("MenuHandler item not found", err, http.StatusNotFound, w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
