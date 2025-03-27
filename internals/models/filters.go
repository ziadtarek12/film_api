package models

import (
	"fmt"
	"math"
	"strings"

	"filmapi.zeyadtarek.net/internals/validator"
)

type Filters struct{
	Page int
	PageSize int
	SortValues []string
	SortSafelist []string
}


type Metadata struct {
	CurrentPage int `json:"current_page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	FirstPage int `json:"first_page,omitempty"`
	LastPage int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func ValidateFilters(v *validator.Validator, filters Filters){
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be a maximum of 100")
	for _, sortValue := range filters.SortValues {
		v.Check(validator.In(sortValue, filters.SortSafelist...), "sort", "invalid sort value: " + sortValue)
	}
}

func (filter *Filters) sortColumn() string{
	sortStr := ""
	invalidSortVal := ""
	for _, sortValue := range filter.SortValues{
		tmp := sortStr
		for _, safeValue := range filter.SortSafelist {
			if sortValue == safeValue {
				sortStr +=  strings.TrimPrefix(sortValue, "-") + sortDirection(sortValue) + ","
			}

			invalidSortVal = sortValue
			
		}
		if tmp == sortStr{
			fmt.Println(tmp)
			fmt.Println(sortStr)
			panic("unsafe sort parameter: " + invalidSortVal)
		}
	}

	return sortStr
}

func sortDirection(sortVal string) string{
	if strings.HasPrefix(sortVal, "-"){
		return " DESC"
	}

	return " ASC"
}

func (filter *Filters) limit() int {
	return filter.PageSize
}

func (filter *Filters) offset() int {
	return (filter.Page - 1) * filter.PageSize
}


func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0{
		return Metadata{}
	}		

	return Metadata{
		CurrentPage: page,
		PageSize: pageSize,
		FirstPage: 1,
		LastPage: int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}