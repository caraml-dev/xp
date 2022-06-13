package controller

import (
	"net/http"

	"github.com/gojek/turing-experiments/common/api/schema"
	api "github.com/gojek/turing-experiments/treatment-service/testhelper/mockmanagement"
	"github.com/gojek/turing-experiments/treatment-service/testhelper/mockmanagement/service"
)

type Segmenter struct {
	SegmenterStore *service.InMemoryStore
}

type SegmenterStore interface {
	ListSegmenters() ([]*schema.Segmenter, error)
	GetSegmenters(projectId int64) ([]*schema.Segmenter, error)
}

func (u Segmenter) GetSegmenters(w http.ResponseWriter, r *http.Request, projectId int64) {
	panic("implement me")
}

func (s Segmenter) ListSegmenters(w http.ResponseWriter, r *http.Request) {
	segmenters, _ := s.SegmenterStore.ListSegmenters()
	response := api.ListSegmentersSuccess{
		Data: segmenters,
	}
	Success(w, response)
}
