package server

import (
	"net/http"
	"net/http/httptest"

	api "github.com/gojek/xp/treatment-service/testhelper/mockmanagement"
	"github.com/gojek/xp/treatment-service/testhelper/mockmanagement/controller"
	"github.com/gojek/xp/treatment-service/testhelper/mockmanagement/service"
)

func newHandler(store *service.InMemoryStore) http.Handler {
	projectSettingsController := controller.ProjectSettings{
		ProjectSettingsStore: store,
	}
	experimentController := controller.Experiment{
		ExperimentStore: store,
	}
	experimentHistoryController := controller.ExperimentHistory{}
	segmenterController := controller.Segmenter{
		SegmenterStore: store,
	}

	return api.Handler(
		controller.NewWrapper(
			projectSettingsController,
			experimentController,
			experimentHistoryController,
			segmenterController,
		),
	)
}

func NewServer(store *service.InMemoryStore) *httptest.Server {
	return httptest.NewServer(newHandler(store))
}
