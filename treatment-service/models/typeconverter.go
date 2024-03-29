package models

import (
	"reflect"
	"time"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	_utils "github.com/caraml-dev/xp/common/utils"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ExperimentTreatment struct {
	Id            *int64                 `json:"id"`
	Configuration map[string]interface{} `json:"configuration"`
	Name          string                 `json:"name"`
	Traffic       *int32                 `json:"traffic,omitempty"`
}

func OpenAPIProjectSettingsSpecToProtobuf(projectSettings schema.ProjectSettings) *_pubsub.ProjectSettings {
	variables := map[string]*_pubsub.ExperimentVariables{}
	for k, v := range projectSettings.Segmenters.Variables.AdditionalProperties {
		experimentVariables := []string{}
		for _, variable := range v {
			experimentVariables = append(experimentVariables, string(variable))
		}
		variables[k] = &_pubsub.ExperimentVariables{Value: experimentVariables}
	}
	segmenters := &_pubsub.Segmenters{
		Names:     projectSettings.Segmenters.Names,
		Variables: variables,
	}

	return &_pubsub.ProjectSettings{
		ProjectId:            projectSettings.ProjectId,
		CreatedAt:            &timestamppb.Timestamp{Seconds: projectSettings.CreatedAt.Unix()},
		UpdatedAt:            &timestamppb.Timestamp{Seconds: projectSettings.UpdatedAt.Unix()},
		Username:             projectSettings.Username,
		Passkey:              projectSettings.Passkey,
		EnableS2IdClustering: projectSettings.EnableS2idClustering,
		Segmenters:           segmenters,
		RandomizationKey:     projectSettings.RandomizationKey,
	}
}

func ProtobufExperimentTypeToOpenAPI(experimentType _pubsub.Experiment_Type) schema.ExperimentType {
	conversionMap := map[_pubsub.Experiment_Type]schema.ExperimentType{
		_pubsub.Experiment_A_B:        schema.ExperimentTypeAB,
		_pubsub.Experiment_Switchback: schema.ExperimentTypeSwitchback,
	}
	return conversionMap[experimentType]
}

func OpenAPIExperimentSpecToProtobuf(
	xpExperiment schema.Experiment,
	segmentersType map[string]schema.SegmenterType,
) (*_pubsub.Experiment, error) {
	statusConverter := map[schema.ExperimentStatus]_pubsub.Experiment_Status{
		"active": _pubsub.Experiment_Active, "inactive": _pubsub.Experiment_Inactive,
	}
	typeConverter := map[schema.ExperimentType]_pubsub.Experiment_Type{
		"Switchback": _pubsub.Experiment_Switchback, "A/B": _pubsub.Experiment_A_B,
	}
	tierConverter := map[schema.ExperimentTier]_pubsub.Experiment_Tier{
		"default": _pubsub.Experiment_Default, "override": _pubsub.Experiment_Override,
	}

	var status _pubsub.Experiment_Status
	if xpExperiment.Status != nil {
		status = statusConverter[*xpExperiment.Status]
	}

	var experimentType _pubsub.Experiment_Type
	if xpExperiment.Type != nil {
		experimentType = typeConverter[*xpExperiment.Type]
	}

	var tier _pubsub.Experiment_Tier
	if xpExperiment.Tier != nil {
		tier = tierConverter[*xpExperiment.Tier]
	}

	segments := make(map[string]*_segmenters.ListSegmenterValue)

	if xpExperiment.Segment != nil {
		for key, val := range *xpExperiment.Segment {
			vals := val.([]interface{})
			switch segmentersType[key] {
			case "string":
				stringVals := []string{}
				for _, val := range vals {
					stringVals = append(stringVals, val.(string))
				}
				segments[key] = _utils.StringSliceToListSegmenterValue(&stringVals)
			case "integer":
				intVals := []int64{}
				for _, val := range vals {
					reflectedVal := reflect.ValueOf(val)
					switch reflectedVal.Kind() {
					case reflect.Float32, reflect.Float64:
						intVals = append(intVals, int64(reflectedVal.Float()))
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						intVals = append(intVals, reflectedVal.Int())
					}
				}
				segments[key] = _utils.Int64ListToListSegmenterValue(&intVals)
			case "real":
				floatVals := []float64{}
				for _, val := range vals {
					floatVals = append(floatVals, val.(float64))
				}
				segments[key] = _utils.FloatListToListSegmenterValue(&floatVals)
			case "bool":
				boolVals := []bool{}
				for _, val := range vals {
					boolVals = append(boolVals, val.(bool))
				}
				segments[key] = _utils.BoolSliceToListSegmenterValue(&boolVals)
			default:
				segments[key] = nil
			}
		}
	}

	treatments := make([]*_pubsub.ExperimentTreatment, 0)
	if xpExperiment.Treatments != nil {
		for _, t := range *xpExperiment.Treatments {
			treatment, err := openAPIExperimentTreatmentSpecToProtobuf(t)
			if err != nil {
				return nil, err
			}
			treatments = append(treatments, treatment)
		}
	}
	interval := int32(0)
	if xpExperiment.Interval != nil {
		interval = *xpExperiment.Interval
	}

	var version int64
	if xpExperiment.Version != nil {
		version = *xpExperiment.Version
	}

	var startTime time.Time
	if xpExperiment.StartTime != nil {
		startTime = *xpExperiment.StartTime
	}

	var endTime time.Time
	if xpExperiment.EndTime != nil {
		endTime = *xpExperiment.EndTime
	}

	var updatedAt time.Time
	if xpExperiment.UpdatedAt != nil {
		updatedAt = *xpExperiment.UpdatedAt
	}

	return &_pubsub.Experiment{
		Id:         *xpExperiment.Id,
		ProjectId:  *xpExperiment.ProjectId,
		Status:     status,
		Name:       *xpExperiment.Name,
		Type:       experimentType,
		Tier:       tier,
		Interval:   interval,
		Segments:   segments,
		Treatments: treatments,
		StartTime:  &timestamppb.Timestamp{Seconds: startTime.Unix()},
		EndTime:    &timestamppb.Timestamp{Seconds: endTime.Unix()},
		UpdatedAt:  &timestamppb.Timestamp{Seconds: updatedAt.Unix()},
		Version:    version,
	}, nil
}

func ExperimentTreatmentToOpenAPITreatment(treatment *_pubsub.ExperimentTreatment) schema.SelectedTreatmentData {
	configuration := DecodeTreatmentConfig(treatment.GetConfig())
	traffic := int32(treatment.GetTraffic())

	selectedTreatment := schema.SelectedTreatmentData{Name: treatment.Name, Traffic: &traffic, Configuration: configuration}

	return selectedTreatment
}

func DecodeTreatmentConfig(config *structpb.Struct) map[string]interface{} {
	if config != nil {
		return config.AsMap()
	}
	return map[string]interface{}{}
}

func openAPIExperimentTreatmentSpecToProtobuf(treatment schema.ExperimentTreatment) (*_pubsub.ExperimentTreatment, error) {
	traffic := uint32(0)
	if treatment.Traffic != nil {
		traffic = uint32(*treatment.Traffic)
	}

	treatmentConfig, err := structpb.NewStruct(treatment.Configuration)
	if err != nil {
		return nil, err
	}

	return &_pubsub.ExperimentTreatment{
		Name:    treatment.Name,
		Config:  treatmentConfig,
		Traffic: traffic,
	}, nil
}
