package routes

import (
	"database/sql"
	"net/http"

	"frappuccino/internal/dal"
	"frappuccino/internal/handler"
	"frappuccino/internal/service"
)

func Routes(mux *http.ServeMux, db *sql.DB) {
	// menu items:
	menuDal := dal.NewMenuRepo(db)
	menuService := service.NewFileMenuService(menuDal)
	menuHandler := handler.NewMenuHandler(menuService)

	mux.HandleFunc("POST /menu", menuHandler.Add)
	mux.HandleFunc("GET /menu", menuHandler.Get)
	mux.HandleFunc("GET /menu/{id}", menuHandler.GetByID)
	mux.HandleFunc("PUT /menu/{id}", menuHandler.Update)
	mux.HandleFunc("DELETE /menu/{id}", menuHandler.Delete)

	// inventory:
	inventoryDal := dal.NewInventoryRepo(db)
	inventoryService := service.NewInventoryService(inventoryDal)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	mux.HandleFunc("POST /inventory", inventoryHandler.Add)
	mux.HandleFunc("GET /inventory", inventoryHandler.Get)
	mux.HandleFunc("GET /inventory/{id}", inventoryHandler.GetByID)
	mux.HandleFunc("PUT /inventory/{id}", inventoryHandler.Update)
	mux.HandleFunc("DELETE /inventory/{id}", inventoryHandler.Delete)
	mux.HandleFunc("GET /inventory/getLeftOvers", inventoryHandler.CountInventory)

	// orders:
	orderDal := dal.NewOrderRepo(db)
	orderService := service.NewOrderService(orderDal)
	orderHandler := handler.NewOrderHandler(orderService)

	mux.HandleFunc("POST /orders", orderHandler.Add)
	mux.HandleFunc("GET /orders", orderHandler.Get)
	mux.HandleFunc("GET /orders/{id}", orderHandler.GetByID)
	mux.HandleFunc("PUT /orders/{id}", orderHandler.Update)
	mux.HandleFunc("DELETE /orders/{id}", orderHandler.Delete)
	mux.HandleFunc("POST /orders/{id}/close", orderHandler.CloseOrder)
	mux.HandleFunc("GET /orders/numberOfOrderedItems", orderHandler.NumberOfOrders)
	mux.HandleFunc("POST /orders/batch-process", orderHandler.BulkOrderProcessing)

	// aggregations:
	reportsDal := dal.NewReportsRepo(db)
	reportsService := service.NewFileReportsService(reportsDal)
	reportsHandler := handler.NewReportsHandler(reportsService)

	mux.HandleFunc("GET /reports/total-sales", reportsHandler.GetTotalSales)
	mux.HandleFunc("GET /reports/popular-items", reportsHandler.GetPopularItems)
	mux.HandleFunc("GET /reports/search", reportsHandler.FullTextSearchReport)
	mux.HandleFunc("GET /reports/orderedItemsByPeriod", reportsHandler.OrderedItemsByPeriod)
}
