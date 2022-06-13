package pagination

import (
	"math"

	"github.com/gojek/turing-experiments/management-service/errors"
)

var (
	MaxPageSize     int32 = 10
	DefaultPageSize int32 = 10
	DefaultPage     int32 = 1
)

type Paging struct {
	Page  int32
	Pages int32
	Total int32
}

type PaginationOptions struct {
	Page     *int32 `json:"page,omitempty"`
	PageSize *int32 `json:"page_size,omitempty"`
}

func NewPaginationOptions(page *int32, pageSize *int32) PaginationOptions {
	if page == nil {
		page = &DefaultPage
	}
	if pageSize == nil {
		pageSize = &DefaultPageSize
	}

	return PaginationOptions{
		Page:     page,
		PageSize: pageSize,
	}
}

func ToPaging(opts PaginationOptions, count int) *Paging {
	return &Paging{
		Page:  *opts.Page,
		Pages: int32(math.Ceil(float64(count) / float64(*opts.PageSize))),
		Total: int32(count),
	}
}

func ValidatePaginationParams(page *int32, pageSize *int32) error {
	if pageSize != nil && (*pageSize <= 0 || *pageSize > MaxPageSize) {
		return errors.Newf(
			errors.BadInput,
			"Received page size. It must be within range (0 < page_size <= %d) or unset.",
			MaxPageSize,
		)
	}
	if page != nil && *page <= 0 {
		return errors.Newf(errors.BadInput, "Received page. It must be > 0 or unset.")
	}

	return nil
}
