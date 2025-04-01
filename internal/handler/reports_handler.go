package handler

import (
	"encoding/json"
	"frapo/internal/service"
	"net/http"
	"strings"
)

type ReportsHandler struct {
	service service.ReportsService
}

func NewReportsHandler(service service.ReportsService) *ReportsHandler {
	return &ReportsHandler{service: service}
}

func (m *ReportsHandler) GetTotalSales(w http.ResponseWriter, r *http.Request) {
	totalSales, err := m.service.TotalPrice()
	if err != nil {
		SendResponse("Failed to get total sales", err, http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(totalSales); err != nil {
		SendResponse("Failed to encode total sales", err, http.StatusInternalServerError, w)
		return
	}
}

func (m *ReportsHandler) GetPopularItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	limit := query.Get("limit")
	if limit == "" {
		limit = "10"
	}
	popularItems, err := m.service.PopularItems(limit)
	if err != nil {
		SendResponse("Failed to get popular items", err, http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(popularItems); err != nil {
		SendResponse("Failed to encode popular items", err, http.StatusInternalServerError, w)
		return
	}
}

func (m *ReportsHandler) FullTextSearchReport(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	q := query.Get("q")

	if q == "" {
		SendResponse("Please write search query", nil, http.StatusInternalServerError, w)
		return
	}

	filter := query.Get("filter")
	filterMap := map[string]bool{}
	if filter == "" {
		filterMap["menu"] = true
		filterMap["orders"] = true
	} else {
		for _, f := range strings.Split(filter, ",") {
			filterMap[strings.TrimSpace(f)] = true
		}
	}

	minPrice := query.Get("minPrice")
	maxPrice := query.Get("maxPrice")

	fullTextSearch, err := m.service.FullTextSearchReport(q, minPrice, maxPrice, filterMap)
	if err != nil {
		SendResponse("Failed to get full text search", err, http.StatusInternalServerError, w)
		return
	}

	w.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(w).Encode(fullTextSearch); err != nil {
		SendResponse("Failed to encode full text search", err, http.StatusInternalServerError, w)
		return
	}
}

func (m *ReportsHandler) OrderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	period := query.Get("period")
	month := query.Get("month")
	year := query.Get("year")

	if period == "day" {
		itemByPeriodMonth, err := m.service.OrderedItemsByPeriodDay(month)
		if err != nil {
			SendResponse("Failed to get item by day", err, http.StatusInternalServerError, w)
			return
		}
		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(itemByPeriodMonth); err != nil {
			SendResponse("Failed to encode ordered items by period", err, http.StatusInternalServerError, w)
			return
		}
	} else if period == "month" {
		itemByPeriodYear, err := m.service.OrderedItemsByPeriodMonth(year)
		if err != nil {
			SendResponse("Failed to get item by month", err, http.StatusInternalServerError, w)
			return
		}
		w.Header().Set("Content-type", "application/json")
		if err := json.NewEncoder(w).Encode(itemByPeriodYear); err != nil {
			SendResponse("Failed to encode ordered items by period", err, http.StatusInternalServerError, w)
			return
		}
	} else {
		SendResponse("Invalid period", nil, http.StatusInternalServerError, w)
		return
	}
}
