package service

import (
	"errors"
	"strconv"

	dal "frappuccino/internal/dal"
	model "frappuccino/models"
)

type ReportsService interface {
	TotalPrice() (model.TotalSalesStruct, error)
	PopularItems(limit string) ([]model.PopularItem, error)
	FullTextSearchReport(q, minPrice, maxPrice string, filterMap map[string]bool) (model.SearchResponse, error)
	OrderedItemsByPeriodDay(month string) (model.ItemByPeriodMonth, error)
	OrderedItemsByPeriodMonth(year string) (model.ItemByPeriodYear, error)
}

type FileReportsService struct {
	repository dal.ReportsDalInterface
}

func NewFileReportsService(repository dal.ReportsDalInterface) *FileReportsService {
	return &FileReportsService{repository: repository}
}

func (f *FileReportsService) TotalPrice() (model.TotalSalesStruct, error) {
	return f.repository.TotalPrice()
}

func (f *FileReportsService) PopularItems(limit string) ([]model.PopularItem, error) {
	return f.repository.PopularItems(limit)
}

func (f *FileReportsService) FullTextSearchReport(q, minPrice, maxPrice string, filterMap map[string]bool) (model.SearchResponse, error) {
	var searchResponse model.SearchResponse

	if minPrice == "" {
		minPrice = "0"
	}

	if maxPrice == "" {
		maxPrice = "9999999999"
	}

	if filterMap["orders"] {
		countOrder, order, err := f.repository.FullTextSearchOrder(q, minPrice, maxPrice)
		if err != nil {
			return model.SearchResponse{}, err
		}
		searchResponse.Orders = order
		searchResponse.TotalMatches += countOrder
	}

	if filterMap["menu"] {
		countMenu, menu, err := f.repository.FullTextSearchMenu(q, minPrice, maxPrice)
		if err != nil {
			return model.SearchResponse{}, err
		}
		searchResponse.MenuItems = menu
		searchResponse.TotalMatches += countMenu
	}

	return searchResponse, nil
}

func (f *FileReportsService) OrderedItemsByPeriodDay(month string) (model.ItemByPeriodMonth, error) {
	monthInt, ok := checkMonth(month)
	if !ok {
		return model.ItemByPeriodMonth{}, errors.New("write month correctly")
	}
	return f.repository.OrderedItemsByPeriodDay(monthInt)
}

func checkMonth(monthCheck string) (int, bool) {
	monthMapping := map[string]int{
		"january":   1,
		"february":  2,
		"march":     3,
		"april":     4,
		"may":       5,
		"june":      6,
		"july":      7,
		"august":    8,
		"september": 9,
		"october":   10,
		"november":  11,
		"december":  12,
	}

	month, ok := monthMapping[monthCheck]
	return month, ok
}

func (f *FileReportsService) OrderedItemsByPeriodMonth(year string) (model.ItemByPeriodYear, error) {
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return model.ItemByPeriodYear{}, err
	}
	return f.repository.OrderedItemsByPeriodMonth(yearInt)
}
