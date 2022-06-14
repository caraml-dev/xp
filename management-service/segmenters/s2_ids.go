package segmenters

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/geo/s2"

	_segmenters "github.com/gojek/xp/common/segmenters"
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

func NewS2IDSegmenter(configData json.RawMessage) (Segmenter, error) {
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
	if config.MinS2CellLevel > config.MaxS2CellLevel {
		return nil, fmt.Errorf(segmenterErrTpl, "Min S2 cell level cannot be greater than max")
	}

	s2IDConfig := &_segmenters.SegmenterConfiguration{
		Name:        "s2_ids",
		Type:        _segmenters.SegmenterValueType_INTEGER,
		Options:     map[string]*_segmenters.SegmenterValue{},
		MultiValued: true,
		TreatmentRequestFields: &_segmenters.ListExperimentVariables{
			Values: []*_segmenters.ExperimentVariables{
				{
					Value: []string{"s2id"},
				},
				{
					Value: []string{"latitude", "longitude"},
				},
			},
		},
		Required:    false,
		Description: fmt.Sprintf("S2 Cell IDs between levels %d and %d are supported.", config.MinS2CellLevel, config.MaxS2CellLevel),
	}

	levels := []int{}
	for i := config.MinS2CellLevel; i <= config.MaxS2CellLevel; i++ {
		levels = append(levels, i)
	}

	return &s2ids{
		NewBaseSegmenter(s2IDConfig),
		levels,
	}, nil
}

type s2ids struct {
	Segmenter
	AllowedLevels []int
}

func (s *s2ids) ValidateSegmenterAndConstraints(segment map[string]*_segmenters.ListSegmenterValue) error {
	err := s.Segmenter.ValidateSegmenterAndConstraints(segment)
	if err != nil {
		return err
	}
	name := s.GetName()

	// Additional check to see that s2id values are valid
	listInputValues := segment[name]
	for _, val := range listInputValues.GetValues() {
		cellID := val.GetInteger()
		cellIDObj := s2.CellID(uint64(cellID))
		if cellID <= 0 || !cellIDObj.IsValid() {
			return fmt.Errorf("One or more %s values is invalid", name)
		} else {
			cellLevel := cellIDObj.Level()
			if !s.isValidLevel(cellLevel) {
				return fmt.Errorf("One or more %s values is at level %d, only the following levels are allowed: %v",
					name, cellLevel, s.AllowedLevels)
			}
		}
	}

	return nil
}

func (s *s2ids) isValidLevel(level int) bool {
	for _, val := range s.AllowedLevels {
		if val == level {
			return true
		}
	}
	return false
}

func init() {
	err := Register("s2_ids", NewS2IDSegmenter)
	if err != nil {
		log.Fatal(err)
	}
}
