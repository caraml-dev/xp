package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
)

type ExperimentTreatments []ExperimentTreatment

type ExperimentTreatment struct {
	Configuration map[string]interface{} `json:"configuration"`
	Name          string                 `json:"name" validate:"required,notBlank"`
	Traffic       *int32                 `json:"traffic,omitempty"`
}

func (t *ExperimentTreatments) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &t)
}

func (t ExperimentTreatments) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t ExperimentTreatments) ToApiSchema() []schema.ExperimentTreatment {
	var treatments []schema.ExperimentTreatment
	for _, treatment := range t {
		treatments = append(treatments, schema.ExperimentTreatment{
			Configuration: treatment.Configuration,
			Name:          treatment.Name,
			Traffic:       treatment.Traffic,
		})
	}
	return treatments
}

func (t ExperimentTreatments) ToProtoSchema() ([]*_pubsub.ExperimentTreatment, error) {
	protoTreatments := make([]*_pubsub.ExperimentTreatment, 0)
	for _, treatment := range t {
		traffic := uint32(0)
		if treatment.Traffic != nil {
			traffic = uint32(*treatment.Traffic)
		}

		treatmentConfig, err := structpb.NewStruct(treatment.Configuration)
		if err != nil {
			return protoTreatments, err
		}

		protoTreatments = append(protoTreatments,
			&_pubsub.ExperimentTreatment{
				Name:    treatment.Name,
				Traffic: traffic,
				Config:  treatmentConfig,
			},
		)
	}
	return protoTreatments, nil
}
