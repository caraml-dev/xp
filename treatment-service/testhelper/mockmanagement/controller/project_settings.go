package controller

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/xp/common/api/schema"
	api "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement"
	"github.com/caraml-dev/xp/treatment-service/util"
)

type ProjectSettings struct {
	ProjectSettingsStore ProjectSettingsStore
}

type ProjectSettingsStore interface {
	ListProjects() ([]schema.Project, error)
	GetProjectSettings(projectId int64) (schema.ProjectSettings, error)
	GetProjectExperimentVariables(projectId int64) ([]string, error)
	UpdateProjectSettings(updated schema.ProjectSettings) error
}

func (u ProjectSettings) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, _ := u.ProjectSettingsStore.ListProjects()
	response := api.ListProjectsSuccess{
		Data: projects,
	}
	Success(w, response)
}

func (u ProjectSettings) GetProjectExperimentVariables(w http.ResponseWriter, r *http.Request, projectId int64) {
	parameter, err := u.ProjectSettingsStore.GetProjectExperimentVariables(projectId)
	if err != nil {
		NotFound(w, err)
	}
	response := api.GetProjectExperimentVariablesSuccess{
		Data: parameter,
	}
	Success(w, response)
}

func (u ProjectSettings) CreateProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	panic("implement me")
}

func (u ProjectSettings) GetProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	settings, err := u.ProjectSettingsStore.GetProjectSettings(projectId)
	if err != nil {
		NotFound(w, err)
		return
	}
	response := api.GetProjectSettingsSuccess{Data: settings}
	Success(w, response)
}

func (u ProjectSettings) UpdateProjectSettings(w http.ResponseWriter, r *http.Request, projectId int64) {
	requestBody := api.UpdateProjectSettingsJSONRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		BadRequest(w, err)
		return
	}
	projectSegmenters := schema.ProjectSegmenters{
		Names:     requestBody.Segmenters.Names,
		Variables: requestBody.Segmenters.Variables,
	}
	updatedProjectSettings := schema.ProjectSettings{
		ProjectId:            projectId,
		EnableS2idClustering: util.DereferenceBool(requestBody.EnableS2idClustering, false),
		RandomizationKey:     requestBody.RandomizationKey,
		Segmenters:           projectSegmenters,
	}
	err = u.ProjectSettingsStore.UpdateProjectSettings(updatedProjectSettings)
	if err != nil {
		BadRequest(w, err)
		return
	}

	response := api.UpdateProjectSettingsSuccess{
		Data: updatedProjectSettings,
	}
	Success(w, response)
}
