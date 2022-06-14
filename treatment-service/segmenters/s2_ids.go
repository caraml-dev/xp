package segmenters

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	_segmenters "github.com/gojek/xp/common/segmenters"
	"github.com/gojek/xp/common/utils"
	"github.com/gojek/xp/treatment-service/util"
)

const (
	// MinS2CellLevel is the permissible minimum value for S2 Cell level
	MinS2CellLevel = 0
	// MaxS2CellLevel is the permissible maximum value for S2 Cell level
	MaxS2CellLevel = 30
)

type S2IDSegmenterConfig struct {
	MinS2CellLevel int `json:"mins2celllevel"`
	MaxS2CellLevel int `json:"maxs2celllevel"`
}

func NewS2IDRunner(configData json.RawMessage) (Runner, error) {
	segmenterErrTpl := "failed to create segmenter (s2_ids): %s"
	var config S2IDSegmenterConfig

	err := json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf(segmenterErrTpl, err)
	}

	// Verify configured levels
	if config.MinS2CellLevel < MinS2CellLevel ||
		config.MinS2CellLevel > MaxS2CellLevel ||
		config.MaxS2CellLevel < MinS2CellLevel ||
		config.MaxS2CellLevel > MaxS2CellLevel {
		return nil, fmt.Errorf(segmenterErrTpl,
			fmt.Sprintf("S2 cell levels should be in the range %d - %d", MinS2CellLevel, MaxS2CellLevel))
	}

	s2IDConfig := &SegmenterConfig{
		Name: "s2_ids",
	}

	return &s2ids{
		NewBaseRunner(s2IDConfig), config.MinS2CellLevel, config.MaxS2CellLevel,
	}, nil
}

type s2ids struct {
	Runner
	MinS2Level int
	MaxS2Level int
}

func (s *s2ids) Transform(
	segmenter string,
	requestValues map[string]interface{},
	experimentVariables []string,
) ([]*_segmenters.SegmenterValue, error) {
	var s2CellID s2.CellID
	switch {
	case cmp.Diff(experimentVariables, []string{"latitude", "longitude"}, cmpopts.SortSlices(utils.Less)) == "":
		err := util.ValidateLatLong(requestValues)
		if err != nil {
			return nil, err
		}
		// Generate S2ID for the supplied level
		latitude, ok := requestValues["latitude"].(float64)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, "latitude", segmenter, "float64")
		}
		longitude, ok := requestValues["longitude"].(float64)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, "longitude", segmenter, "float64")
		}
		retrievedS2id, err := util.GetS2ID(latitude, longitude, s.MaxS2Level)
		if err != nil {
			return nil, err
		}
		s2CellID = retrievedS2id
	case cmp.Equal(experimentVariables, []string{"s2id"}):
		s2idCast, ok := requestValues["s2id"].(float64)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, "s2id", segmenter, "float64")
		}
		s2CellID = s2.CellID(int64(s2idCast))
		if !s2CellID.IsValid() {
			return nil, fmt.Errorf("provided s2id variable for %s segmenter is invalid", segmenter)
		}
	default:
		return nil, fmt.Errorf("no valid variables were provided for %s segmenter", segmenter)
	}
	segmenterValues := []*_segmenters.SegmenterValue{}

	// Order defines S2ID matching priority, i.e match S2ID based on decreasing granularity
	for i := s.MaxS2Level; i >= s.MinS2Level; i-- {
		s2IdAtLevel := int64(s2CellID.Parent(i))
		segmenterValue := &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: s2IdAtLevel}}
		segmenterValues = append(segmenterValues, segmenterValue)
	}

	return segmenterValues, nil
}

func init() {
	err := Register("s2_ids", NewS2IDRunner)
	if err != nil {
		log.Fatal(err)
	}
}
