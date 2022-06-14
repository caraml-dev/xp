package services

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang/geo/s2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gojek/xp/common/api/schema"
	_pubsub "github.com/gojek/xp/common/pubsub"
	_segmenters "github.com/gojek/xp/common/segmenters"
	tu "github.com/gojek/xp/common/testutils"
	"github.com/gojek/xp/treatment-service/models"
)

type ExperimentServiceTestSuite struct {
	suite.Suite
	ExperimentService
	// LocalStorage holds the data created during the setup, used by the tests
	LocalStorage models.LocalStorage
}

func (s *ExperimentServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ExperimentServiceTestSuite")

	// Define segmenters and segments
	// S2IDs - all S2IDs below contain the one used in the test request params
	s2ID1 := interface{}(int64(3592210809859604480)) // Level 14
	s2ID2 := interface{}(int64(3592214439106969600)) // Level 9
	s2ID3 := interface{}(int64(3592210796974702592)) // Level 12

	// Other segmenters
	// Casting to float64 because JSON Unmarshall treat numbers as floats
	rawStringSegmenter := interface{}("seg-1")
	rawIntegerSegmenter := interface{}(int64(1))
	rawFloatSegmenter := interface{}(float64(9001))
	rawBoolSegmenter := interface{}(false)
	daysOfWeek := interface{}(int64(1))
	hoursOfDay := interface{}(int64(20))
	segment1 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, &rawIntegerSegmenter, &rawFloatSegmenter, nil)
	segment2 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, &rawIntegerSegmenter, &rawFloatSegmenter, &s2ID2)
	segment3 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, &rawIntegerSegmenter, &rawFloatSegmenter, &s2ID3)
	segment4 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, nil, &rawFloatSegmenter, &s2ID3)
	segment5 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, &rawIntegerSegmenter, nil, &s2ID3)
	segment6 := makeSegment(
		&rawStringSegmenter, &daysOfWeek, &hoursOfDay, &rawBoolSegmenter, &rawIntegerSegmenter, nil, &s2ID1)

	// Set up local storage
	s.LocalStorage = models.LocalStorage{
		ProjectSettings: []*_pubsub.ProjectSettings{
			{ProjectId: 1},
			{ProjectId: 2, Segmenters: &_pubsub.Segmenters{}},
			{ProjectId: 3, Segmenters: &_pubsub.Segmenters{}},
			{
				ProjectId: 4,
				// Segmenter order influences results
				Segmenters: &_pubsub.Segmenters{
					Names: []string{
						"string_segmenter", "integer_segmenter", "s2_ids", "float_segmenter", "days_of_week",
						"hours_of_day", "bool_segmenter",
					},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"string_segmenter":  {Value: []string{"string_segmenter"}},
						"integer_segmenter": {Value: []string{"integer_segmenter"}},
						"s2_ids":            {Value: []string{"latitude", "longitude"}},
						"float_segmenter":   {Value: []string{"float_segmenter"}},
						"days_of_week":      {Value: []string{"tz"}},
						"hours_of_day":      {Value: []string{"tz"}},
						"bool_segmenter":    {Value: []string{"bool_segmenter"}},
					},
				},
			},
			{ProjectId: 5, Segmenters: &_pubsub.Segmenters{}},
			{
				ProjectId: 6,
				// Segmenter order influences results
				Segmenters: &_pubsub.Segmenters{
					Names: []string{
						"string_segmenter", "float_segmenter", "integer_segmenter", "s2_ids", "days_of_week",
						"hours_of_day", "bool_segmenter",
					},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"string_segmenter":  {Value: []string{"string_segmenter"}},
						"float_segmenter":   {Value: []string{"float_segmenter"}},
						"integer_segmenter": {Value: []string{"integer_segmenter"}},
						"s2_ids":            {Value: []string{"latitude", "longitude"}},
						"days_of_week":      {Value: []string{"tz"}},
						"hours_of_day":      {Value: []string{"tz"}},
						"bool_segmenter":    {Value: []string{"bool_segmenter"}},
					},
				},
			},
			{
				ProjectId: 7,
				// Segmenter order influences results
				Segmenters: &_pubsub.Segmenters{
					Names: []string{
						"string_segmenter", "integer_segmenter", "s2_ids", "days_of_week",
						"hours_of_day", "bool_segmenter",
					},
					Variables: map[string]*_pubsub.ExperimentVariables{
						"string_segmenter":  {Value: []string{"string_segmenter"}},
						"integer_segmenter": {Value: []string{"integer_segmenter"}},
						"float_segmenter":   {Value: []string{"float_segmenter"}},
						"s2_ids":            {Value: []string{"latitude", "longitude"}},
						"days_of_week":      {Value: []string{"tz"}},
						"hours_of_day":      {Value: []string{"tz"}},
						"bool_segmenter":    {Value: []string{"bool_segmenter"}},
					},
				},
			},
		},
		Experiments: map[models.ProjectId][]*models.ExperimentIndex{
			1: {makeExperimentIndex(1, 1, segment1, _pubsub.Experiment_Default)},
			// Experiments contain the same data
			2: {
				makeExperimentIndex(2, 1, segment1, _pubsub.Experiment_Default),
				makeExperimentIndex(2, 2, segment1, _pubsub.Experiment_Default),
			},
			// Experiments contain s2IDs at different levels
			3: {
				makeExperimentIndex(3, 1, segment2, _pubsub.Experiment_Default),
				makeExperimentIndex(3, 2, segment3, _pubsub.Experiment_Default),
			},
			// Experiments contain optional segmenters
			4: {
				makeExperimentIndex(4, 1, segment4, _pubsub.Experiment_Default),
				makeExperimentIndex(4, 2, segment5, _pubsub.Experiment_Default),
			},
			// Experiments contain varying tiers, no S2ID
			5: {
				makeExperimentIndex(5, 1, segment1, _pubsub.Experiment_Override),
				makeExperimentIndex(5, 2, segment1, _pubsub.Experiment_Default),
			},
			// FloatSegmenter experiment of highest priority
			6: {
				makeExperimentIndex(6, 1, segment4, _pubsub.Experiment_Default),
				makeExperimentIndex(6, 2, segment5, _pubsub.Experiment_Default),
				makeExperimentIndex(6, 3, segment6, _pubsub.Experiment_Default),
			},
			// FloatSegmenter + granular S2ID experiment of highest priority
			7: {
				makeExperimentIndex(7, 1, segment4, _pubsub.Experiment_Default),
				makeExperimentIndex(7, 2, segment5, _pubsub.Experiment_Default),
				makeExperimentIndex(7, 3, segment5, _pubsub.Experiment_Override),
				makeExperimentIndex(7, 4, segment6, _pubsub.Experiment_Default),
			},
		},
		Segmenters: map[string]schema.SegmenterType{
			"string_segmenter":  "STRING",
			"integer_segmenter": "INTEGER",
			"float_segmenter":   "FLOAT",
			"bool_segmenter":    "BOOL",
			"s2_ids":            "INTEGER",
			"days_of_week":      "INTEGER",
			"hours_of_day":      "INTEGER",
		},
	}

	var err error
	s.ExperimentService, err = NewExperimentService(
		&s.LocalStorage,
	)
	if err != nil {
		s.Suite.T().Fatalf("Could not start up ExperimentService: %v", err)
	}
}

func (s *ExperimentServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up ExperimentServiceTestSuite")
}

func TestExperimentService(t *testing.T) {
	suite.Run(t, new(ExperimentServiceTestSuite))
}

func (s *ExperimentServiceTestSuite) TestGetExperiment() {
	tests := map[string]struct {
		description      string // Optional - additional description about the test
		projectId        uint32
		reqFilter        map[string][]*_segmenters.SegmenterValue
		expLookupFilters []models.SegmentFilter
		expResponse      *_pubsub.Experiment
		expError         string
	}{
		"no experiment | segment filter unmatched": {
			description:      "Bool segmenter filter doesn't match any experiment",
			projectId:        1,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, true),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "SG", 1, 9001, true),
		},
		"no experiment | location unmatched": {
			description:      "S2ID filter doesn't match any experiment",
			projectId:        4,
			reqFilter:        makeRequestFilter(s2.CellID(3591916853707931648), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3591916853707931648), 1, 20, "seg-1", 1, 9001, false),
		},
		"single experiment": {
			projectId:        1,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[1][0].Experiment,
		},
		"multiple experiments error": {
			projectId:        2,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expError:         "more than 1 experiment of the same match strength encountered",
		},
		"resolve variable s2IDs": {
			projectId:        3,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[3][1].Experiment,
		},
		"resolve optional segmenters": {
			projectId:        4,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[4][1].Experiment,
		},
		"resolve tiers": {
			projectId:        5,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[5][0].Experiment,
		},
		"resolve all hierarchy | optional segmenter": {
			projectId:        6,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[6][0].Experiment,
		},
		"resolve all hierarchy | s2id granularity": {
			projectId:        7,
			reqFilter:        makeRequestFilter(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expLookupFilters: makeExperimentLookupFilters(s2.CellID(3592210809859604480), 1, 20, "seg-1", 1, 9001, false),
			expResponse:      s.LocalStorage.Experiments[7][3].Experiment,
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			expLookup, expResponse, err := s.ExperimentService.GetExperiment(tt.projectId, tt.reqFilter)

			assert.Equal(t, tt.expResponse, expResponse)
			assert.True(t, true, reflect.DeepEqual(tt.expLookupFilters, expLookup))
			if tt.expError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expError)
			}
		})
	}
}

func (s *ExperimentServiceTestSuite) TestDumpExperiments() {
	filename, err := s.ExperimentService.DumpExperiments("/tmp")
	s.Suite.T().Log(filename)
	// Defer file cleanup
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			s.Suite.T().Logf("Error cleaning up file: %s", filename)
		}
	}()
	s.Suite.Require().NoError(err)
	// Read the result file and unmarshal the JSON content
	bytes, err := tu.ReadFile(filename)
	s.Suite.Require().NoError(err)
	var results map[string][]models.ExperimentIndexLog
	err = json.Unmarshal(bytes, &results)
	s.Suite.Require().NoError(err)
	// Test some properties of the results.
	// (The full results are tests in the storage tests.)
	projectIds := []string{}
	experimentCount := map[string]int{
		"1": 1, "2": 2, "3": 2, "4": 2, "5": 2, "6": 3, "7": 4,
	}
	for k, v := range results {
		projectIds = append(projectIds, k)
		if count, ok := experimentCount[k]; ok {
			s.Suite.Assert().Equal(count, len(v))
		}
	}
	s.Suite.Assert().Equal(7, len(projectIds))
}

func makeExperimentIndex(
	projectId int64,
	id int64,
	segment map[string]*_segmenters.ListSegmenterValue,
	tier _pubsub.Experiment_Tier,
) *models.ExperimentIndex {
	hourStart := time.Now().Truncate(time.Hour)
	hourEnd := time.Now().Truncate(time.Hour).Add(time.Hour * 2)

	return models.NewExperimentIndex(&_pubsub.Experiment{
		Id:        id,
		ProjectId: projectId,
		Name:      fmt.Sprintf("exp-%d", id),
		Segments:  segment,
		Tier:      tier,
		Type:      _pubsub.Experiment_A_B,
		Status:    _pubsub.Experiment_Active,
		StartTime: &timestamppb.Timestamp{Seconds: hourStart.Unix()},
		EndTime:   &timestamppb.Timestamp{Seconds: hourEnd.Unix()},
	})
}

func makeSegment(
	rawStringSegmenter *interface{},
	daysOfWeek *interface{},
	hoursOfDay *interface{},
	rawBoolSegmenter *interface{},
	rawIntegerSegmenter *interface{},
	rawFloatSegmenter *interface{},
	s2Ids *interface{},
) map[string]*_segmenters.ListSegmenterValue {
	segments := make(map[string]*_segmenters.ListSegmenterValue)
	if rawStringSegmenter != nil {
		stringSegmenter := *rawStringSegmenter
		segments["string_segmenter"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_String_{String_: stringSegmenter.(string)}},
			},
		}
	}
	if rawBoolSegmenter != nil {
		boolSegmenter := *rawBoolSegmenter
		segments["bool_segmenter"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Bool{Bool: boolSegmenter.(bool)}},
			},
		}
	}
	if rawFloatSegmenter != nil {
		floatSegmenter := *rawFloatSegmenter
		segments["float_segmenter"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Real{Real: floatSegmenter.(float64)}},
			},
		}
	}
	if rawIntegerSegmenter != nil {
		integerSegmenter := *rawIntegerSegmenter
		segments["integer_segmenter"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: integerSegmenter.(int64)}},
			},
		}
	}
	if s2Ids != nil {
		s2_ids := *s2Ids
		segments["s2_ids"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: s2_ids.(int64)}},
			},
		}
	}
	if daysOfWeek != nil {
		dow := *daysOfWeek
		segments["days_of_week"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: dow.(int64)}},
			},
		}
	}
	if hoursOfDay != nil {
		hod := *hoursOfDay
		segments["hours_of_day"] = &_segmenters.ListSegmenterValue{
			Values: []*_segmenters.SegmenterValue{
				{Value: &_segmenters.SegmenterValue_Integer{Integer: hod.(int64)}},
			},
		}
	}

	return segments
}

func makeRequestFilter(
	s2ID s2.CellID,
	daysOfWeek int64,
	hoursOfDay int64,
	stringSegmenter string,
	integerSegmenter int64,
	floatSegmenter float64,
	boolSegmenter bool,
) map[string][]*_segmenters.SegmenterValue {
	segmenterValues := []*_segmenters.SegmenterValue{}
	for i := 14; i >= 10; i-- {
		s2IdAtLevel := int64(s2ID.Parent(i))
		segmenterValue := &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: s2IdAtLevel}}
		segmenterValues = append(segmenterValues, segmenterValue)
	}
	return map[string][]*_segmenters.SegmenterValue{
		"string_segmenter":  {{Value: &_segmenters.SegmenterValue_String_{String_: stringSegmenter}}},
		"integer_segmenter": {{Value: &_segmenters.SegmenterValue_Integer{Integer: integerSegmenter}}},
		"float_segmenter":   {{Value: &_segmenters.SegmenterValue_Real{Real: floatSegmenter}}},
		"bool_segmenter":    {{Value: &_segmenters.SegmenterValue_Bool{Bool: boolSegmenter}}},
		"s2_ids":            segmenterValues,
		"days_of_week":      {{Value: &_segmenters.SegmenterValue_Integer{Integer: daysOfWeek}}},
		"hours_of_day":      {{Value: &_segmenters.SegmenterValue_Integer{Integer: hoursOfDay}}},
	}
}

func makeExperimentLookupFilters(
	s2ID s2.CellID,
	daysOfWeek int64,
	hoursOfDay int64,
	stringSegmenter string,
	integerSegmenter int64,
	floatSegmenter float64,
	boolSegmenter bool,
) []models.SegmentFilter {
	segmenterValues := []*_segmenters.SegmenterValue{}
	for i := 14; i >= 10; i-- {
		s2IdAtLevel := int64(s2ID.Parent(i))
		segmenterValue := &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: s2IdAtLevel}}
		segmenterValues = append(segmenterValues, segmenterValue)
	}
	return []models.SegmentFilter{
		{Key: "string_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: stringSegmenter}}}},
		{Key: "float_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Real{Real: floatSegmenter}}}},
		{Key: "integer_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: integerSegmenter}}}},
		{Key: "bool_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: boolSegmenter}}}},
		{Key: "s2_ids", Value: segmenterValues},
		{Key: "days_of_week", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: daysOfWeek}}}},
		{Key: "hours_of_day", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: hoursOfDay}}}},
	}
}
