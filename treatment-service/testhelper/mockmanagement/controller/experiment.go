package controller

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/xp/common/api/schema"
	api "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement"
	"github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement/service"
)

type Experiment struct {
	ExperimentStore *service.InMemoryStore
}

func (e Experiment) GetExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	experiment, err := e.ExperimentStore.GetExperiment(projectId, experimentId)
	if err != nil {
		NotFound(w, err)
		return
	}
	response := api.GetExperimentSuccess{Data: experiment}
	Success(w, response)
}

func (e Experiment) ListExperiments(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListExperimentsParams) {
	experiments, _ := e.ExperimentStore.ListExperiments(projectId, params)
	response := api.ListExperimentsSuccess{
		Data: experiments,
	}
	Success(w, response)
}

func (e Experiment) CreateExperiment(w http.ResponseWriter, r *http.Request, projectId int64) {
	requestBody := api.CreateExperimentJSONRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		BadRequest(w, err)
		return
	}
	experiment := schema.Experiment{
		ProjectId:   projectId,
		Description: requestBody.Description,
		EndTime:     requestBody.EndTime,
		Interval:    requestBody.Interval,
		Name:        requestBody.Name,
		Segment:     requestBody.Segment,
		StartTime:   requestBody.StartTime,
		Status:      requestBody.Status,
		Treatments:  requestBody.Treatments,
		Type:        requestBody.Type,
		UpdatedBy:   *requestBody.UpdatedBy,
	}
	createdExperiment, err := e.ExperimentStore.CreateExperiment(experiment)
	if err != nil {
		BadRequest(w, err)
		return
	}
	response := api.CreateExperimentSuccess{Data: createdExperiment}
	Success(w, response)
}

func (e Experiment) UpdateExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	experiment, err := e.ExperimentStore.GetExperiment(projectId, experimentId)
	if err != nil {
		NotFound(w, err)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&experiment)
	if err != nil {
		BadRequest(w, err)
		return
	}
	updatedExperiment, err := e.ExperimentStore.UpdateExperiment(projectId, experimentId, experiment)
	if err != nil {
		BadRequest(w, err)
		return
	}
	response := api.UpdateExperimentSuccess{Data: updatedExperiment}
	Success(w, response)
}

func (e Experiment) EnableExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	experiment, err := e.ExperimentStore.GetExperiment(projectId, experimentId)
	if err != nil {
		NotFound(w, err)
		return
	}
	experiment.Status = schema.ExperimentStatusActive
	updatedExperiment, err := e.ExperimentStore.UpdateExperiment(projectId, experimentId, experiment)
	if err != nil {
		BadRequest(w, err)
		return
	}
	response := api.UpdateExperimentSuccess{Data: updatedExperiment}
	Success(w, response)
}

func (e Experiment) DisableExperiment(w http.ResponseWriter, r *http.Request, projectId int64, experimentId int64) {
	experiment, err := e.ExperimentStore.GetExperiment(projectId, experimentId)
	if err != nil {
		NotFound(w, err)
		return
	}
	experiment.Status = schema.ExperimentStatusInactive
	updatedExperiment, err := e.ExperimentStore.UpdateExperiment(projectId, experimentId, experiment)
	if err != nil {
		BadRequest(w, err)
		return
	}
	response := api.UpdateExperimentSuccess{Data: updatedExperiment}
	Success(w, response)
}
