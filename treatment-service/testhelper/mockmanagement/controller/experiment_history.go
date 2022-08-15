package controller

import (
	"net/http"

	api "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement"
)

type ExperimentHistory struct{}

func (e ExperimentHistory) ListExperimentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	experimentId int64,
	params api.ListExperimentHistoryParams,
) {
}

func (e ExperimentHistory) GetExperimentHistory(
	w http.ResponseWriter,
	r *http.Request,
	projectId int64,
	experimentId int64,
	version int64) {
}
