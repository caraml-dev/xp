package pagination

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPaginationOptions(t *testing.T) {
	one, two, three, ten := int32(1), int32(2), int32(3), int32(10)

	tests := map[string]struct {
		page     *int32
		pageSize *int32
		expected PaginationOptions
	}{
		"defaults": {
			expected: PaginationOptions{
				Page:     &one,
				PageSize: &ten,
			},
		},
		"missing page": {
			pageSize: &three,
			expected: PaginationOptions{
				Page:     &one,
				PageSize: &three,
			},
		},
		"missing page size": {
			page: &two,
			expected: PaginationOptions{
				Page:     &two,
				PageSize: &ten,
			},
		},
		"new values": {
			page:     &two,
			pageSize: &three,
			expected: PaginationOptions{
				Page:     &two,
				PageSize: &three,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, data.expected, NewPaginationOptions(data.page, data.pageSize))
		})
	}
}

func TestPaging(t *testing.T) {
	tests := []struct {
		page     int32
		pageSize int32
		count    int
		expected Paging
	}{
		{
			page:     int32(1),
			pageSize: int32(10),
			count:    5,
			expected: Paging{
				Page:  1,
				Pages: 1,
				Total: 5,
			},
		},
		{
			page:     int32(2),
			pageSize: int32(3),
			count:    7,
			expected: Paging{
				Page:  2,
				Pages: 3,
				Total: 7,
			},
		},
	}

	for idx, data := range tests {
		t.Run(fmt.Sprintf("case %d", idx), func(t *testing.T) {
			actual := ToPaging(PaginationOptions{Page: &data.page, PageSize: &data.pageSize}, data.count)
			assert.Equal(t, data.expected, *actual)
		})
	}
}
