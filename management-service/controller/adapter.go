package controller

import (
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/pagination"
)

func ToPagingSchema(p *pagination.Paging) *schema.Paging {
	var paging schema.Paging
	if p == nil {
		return nil
	}
	paging = schema.Paging(*p)
	return &paging
}
