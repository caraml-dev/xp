package controller

import (
	"net/http"

	"github.com/caraml-dev/xp/common/api/schema"
	api "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement"
	"github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement/service"
)

type Segmenter struct {
	SegmenterStore *service.InMemoryStore
}

type SegmenterStore interface {
	ListSegmenters(projectId int64) ([]*schema.Segmenter, error)
}

func (s Segmenter) ListSegmenters(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListSegmentersParams) {
	segmenters, _ := s.SegmenterStore.ListSegmenters(projectId)
	response := api.ListSegmentersSuccess{
		Data: segmenters,
	}
	Success(w, response)
}

func (s Segmenter) GetSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	panic("implement me")
}

func (s Segmenter) CreateSegmenter(w http.ResponseWriter, r *http.Request, projectId int64) {
	panic("implement me")
}

func (s Segmenter) UpdateSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	panic("implement me")
}

func (s Segmenter) DeleteSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	panic("implement me")
}
