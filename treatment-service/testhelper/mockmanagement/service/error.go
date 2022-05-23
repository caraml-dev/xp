package service

import "fmt"

type InvalidExperiment struct {
	projectId    int64
	experimentId int64
}

func (i InvalidExperiment) Error() string {
	return fmt.Sprintf("experiment id %d does not exist for project id %d", i.experimentId, i.projectId)
}

type InvalidProjectSettings struct {
	projectId int64
}

func (i InvalidProjectSettings) Error() string {
	return fmt.Sprintf("setting for project id %d does not exist", i.projectId)
}

type InvalidTreatment struct {
	projectId   int64
	treatmentId int64
}

func (i InvalidTreatment) Error() string {
	return fmt.Sprintf("treatment id %d does not exist for project id %d", i.treatmentId, i.projectId)
}
