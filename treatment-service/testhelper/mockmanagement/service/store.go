package service

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-collections/collections/set"

	"github.com/caraml-dev/xp/common/api/schema"
	api "github.com/caraml-dev/xp/treatment-service/testhelper/mockmanagement"
)

type InMemoryStore struct {
	sync.RWMutex
	Experiments     []schema.Experiment
	ProjectSettings []schema.ProjectSettings
	MessageQueue    MessageQueue
	SegmentersTypes map[string]schema.SegmenterType
}

func NewInMemoryStore(
	experiments []schema.Experiment,
	settings []schema.ProjectSettings,
	queue MessageQueue,
	segmentersType map[string]schema.SegmenterType,
) (*InMemoryStore, error) {
	return &InMemoryStore{
		Experiments:     experiments,
		ProjectSettings: settings,
		MessageQueue:    queue,
		SegmentersTypes: segmentersType,
	}, nil
}

func (i *InMemoryStore) GetExperiment(projectId int64, experimentId int64) (schema.Experiment, error) {
	i.RLock()
	defer i.RUnlock()
	for _, experiment := range i.Experiments {
		if experiment.ProjectId == projectId && experiment.Id == experimentId {
			return experiment, nil
		}
	}
	return schema.Experiment{}, InvalidExperiment{
		projectId:    projectId,
		experimentId: experimentId,
	}
}

func (i *InMemoryStore) ListExperiments(projectId int64, params api.ListExperimentsParams) ([]schema.Experiment, error) {
	i.RLock()
	defer i.RUnlock()
	projectExperiments := make([]schema.Experiment, 0)
	for _, experiment := range i.Experiments {
		if experiment.ProjectId == projectId {
			projectExperiments = append(projectExperiments, experiment)
		}
	}
	return projectExperiments, nil
}

func (i *InMemoryStore) CreateExperiment(experiment schema.Experiment) (schema.Experiment, error) {
	i.Lock()
	defer i.Unlock()
	experiment.Id = int64(len(i.Experiments)) + 1
	experiment.Status = schema.ExperimentStatusActive
	experiment.UpdatedAt = time.Now()
	i.Experiments = append(i.Experiments, experiment)

	err := i.MessageQueue.PublishNewExperiment(experiment, i.SegmentersTypes)
	if err != nil {
		return schema.Experiment{}, err
	}

	return experiment, nil
}

func (i *InMemoryStore) UpdateExperiment(projectId int64, experimentId int64, experiment schema.Experiment) (schema.Experiment, error) {
	i.Lock()
	defer i.Unlock()
	experiment.UpdatedAt = time.Now()
	for index, experiment := range i.Experiments {
		if experiment.ProjectId == projectId && experiment.Id == experimentId {
			i.Experiments[index] = experiment
		}
	}

	err := i.MessageQueue.UpdateExperiment(experiment, i.SegmentersTypes)
	if err != nil {
		return schema.Experiment{}, err
	}

	return experiment, nil
}

func (i *InMemoryStore) ListProjects() ([]schema.Project, error) {
	projects := []schema.Project{}
	for _, settings := range i.ProjectSettings {
		projects = append(projects, schema.Project{
			Id:               settings.ProjectId,
			CreatedAt:        settings.CreatedAt,
			UpdatedAt:        settings.UpdatedAt,
			Username:         settings.Username,
			RandomizationKey: settings.RandomizationKey,
			Segmenters:       settings.Segmenters.Names,
		})
	}
	return projects, nil
}

func (i *InMemoryStore) GetProjectExperimentVariables(projectId int64) ([]string, error) {
	i.RLock()
	defer i.RUnlock()
	var settings schema.ProjectSettings
	settingsPresent := false
	for _, val := range i.ProjectSettings {
		if val.ProjectId == projectId {
			settings = val
			settingsPresent = true
			break
		}
	}
	if !settingsPresent {
		return nil, errors.New("project id does not exist")
	}

	return ConvertExperimentVariablesType(settings.Segmenters.Variables), nil
}

func ConvertExperimentVariablesType(parameters schema.ProjectSegmenters_Variables) []string {
	var requestParams []string
	segmenterSet := set.New()
	for _, val := range parameters.AdditionalProperties {
		for _, variable := range val {
			if segmenterSet.Has(variable) {
				continue
			}
			segmenterSet.Insert(variable)
			requestParams = append(requestParams, variable)
		}
	}
	return requestParams
}

func (i *InMemoryStore) UpdateProjectSettings(updated schema.ProjectSettings) error {
	i.Lock()
	defer i.Unlock()
	for index, settings := range i.ProjectSettings {
		if settings.ProjectId == updated.ProjectId {
			i.ProjectSettings[index] = updated
			return i.MessageQueue.UpdateProjectSettings(updated)
		}
	}
	return InvalidProjectSettings{
		projectId: updated.ProjectId,
	}
}

func (i *InMemoryStore) GetProjectSettings(projectId int64) (schema.ProjectSettings, error) {
	i.Lock()
	defer i.Unlock()
	for _, settings := range i.ProjectSettings {
		if settings.ProjectId == projectId {
			return settings, nil
		}
	}
	return schema.ProjectSettings{}, InvalidProjectSettings{projectId: projectId}
}

// TODO to be implemented in next MR
func (i *InMemoryStore) ListSegmenters(projectId int64) ([]schema.Segmenter, error) {
	i.RLock()
	defer i.RUnlock()
	projectSegmenters := make([]schema.Segmenter, 0)
	for segmenterName, segmenterType := range i.SegmentersTypes {
		segmenter := schema.Segmenter{Name: segmenterName, Type: segmenterType}
		projectSegmenters = append(projectSegmenters, segmenter)
	}
	return projectSegmenters, nil
}
