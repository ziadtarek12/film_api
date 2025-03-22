package models

import "filmapi.zeyadtarek.net/internals/validator"

type Filters struct{
	Page int
	PageSize int
	Sort string
	SortSafelist []string
}

func ValidateFilters(v *validator.Validator, filters Filters){
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be a maximum of 100")
	v.Check(validator.In(filters.Sort, filters.SortSafelist...), "sort", "invalid sort value")
}