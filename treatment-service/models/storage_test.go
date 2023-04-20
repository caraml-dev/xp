package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-collections/collections/set"
	"github.com/golang/geo/s2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	managementClient "github.com/caraml-dev/xp/clients/management"
	mocks "github.com/caraml-dev/xp/clients/testutils/mocks/management"
	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	tu "github.com/caraml-dev/xp/common/testutils"
)

type LocalStorageLookupSuite struct {
	suite.Suite

	testExperiments []*_pubsub.Experiment
	storage         LocalStorage
	location        s2.CellID
}

func newTestXPExperiment(
	projectId int64,
	segment schema.ExperimentSegment,
	startTime time.Time,
	endTime time.Time,
) schema.Experiment {
	name, _ := uuid.NewUUID()
	interval := int32(30)
	traffic := int32(100)
	treatments := []schema.ExperimentTreatment{
		{Configuration: make(map[string]interface{}), Name: "default", Traffic: &traffic},
	}

	id := int64(rand.Intn(100))
	nameString := name.String()
	status := schema.ExperimentStatusActive
	experimentType := schema.ExperimentTypeAB
	updatedAt := time.Time{}
	updatedBy := ""

	return schema.Experiment{
		ProjectId:  &projectId,
		Id:         &id,
		StartTime:  &startTime,
		EndTime:    &endTime,
		Name:       &nameString,
		Status:     &status,
		Segment:    &segment,
		Treatments: &treatments,
		Type:       &experimentType,
		UpdatedAt:  &updatedAt,
		UpdatedBy:  &updatedBy,
		Interval:   &interval,
	}
}

func newProjectSettings(
	enableS2idClustering bool,
	passkey string,
	projectId int64,
	randomizationKey string,
	segmenters []string,
	username string,
) schema.ProjectSettings {
	createdAt := time.Now()

	variablesProperties := map[string][]string{}
	for _, segmenter := range segmenters {
		variablesProperties[segmenter] = []string{segmenter}
	}
	formattedSegmenters := schema.ProjectSegmenters{
		Names:     segmenters,
		Variables: schema.ProjectSegmenters_Variables{AdditionalProperties: variablesProperties},
	}

	return schema.ProjectSettings{
		CreatedAt:            createdAt,
		EnableS2idClustering: enableS2idClustering,
		Passkey:              passkey,
		ProjectId:            projectId,
		RandomizationKey:     randomizationKey,
		Segmenters:           formattedSegmenters,
		UpdatedAt:            createdAt,
		Username:             username,
	}
}

func (suite *LocalStorageLookupSuite) SetupTest() {
	projectId := uint32(0)

	mockManagementClientInterface := mocks.ClientInterface{}
	mockManagementClient := managementClient.ClientWithResponses{ClientInterface: &mockManagementClientInterface}
	mockManagementClientInterface.On("ListSegmenters",
		context.TODO(),
		int64(3),
		&managementClient.ListSegmentersParams{}).
		Return(&http.Response{
			StatusCode: 200,
			Header:     map[string][]string{"Content-Type": {"json"}},
			Body:       io.NopCloser(bytes.NewBufferString(`{"data" : []}`)),
		}, nil)

	suite.storage = LocalStorage{
		Experiments:       make(map[ProjectId][]*ExperimentIndex),
		managementClient:  &mockManagementClient,
		ProjectSegmenters: map[ProjectId]map[string]schema.SegmenterType{},
	}
	suite.storage.Experiments[projectId] = make([]*ExperimentIndex, 0)
	suite.storage.ProjectSettings = []*_pubsub.ProjectSettings{}

	suite.testExperiments = make([]*_pubsub.Experiment, 0)
	segmentersType := map[string]schema.SegmenterType{
		"string_segmenter":    "string",
		"integer_segmenter":   "integer",
		"integer_segmenter_2": "integer",
		"bool_segmenter":      "bool",
		"s2_ids":              "integer",
	}

	addExperiment := func(experiment schema.Experiment) {
		e, err := OpenAPIExperimentSpecToProtobuf(experiment, segmentersType)
		suite.Require().NoError(err)
		suite.testExperiments = append(suite.testExperiments, e)
		suite.storage.Experiments[projectId] = append(suite.storage.Experiments[projectId], NewExperimentIndex(e))
	}
	addProjectSettings := func(projectSettings schema.ProjectSettings) {
		e := OpenAPIProjectSettingsSpecToProtobuf(projectSettings)
		suite.storage.ProjectSettings = append(suite.storage.ProjectSettings, e)
	}

	// Add Projects
	addProjectSettings(newProjectSettings(
		false, "passkey", 1, "randomkey", []string{"string_segmenter", "integer_segmenter", "integer_segmenter_2"}, "user1"))
	addProjectSettings(newProjectSettings(
		false, "passkey", 2, "randomkey", []string{"string_segmenter"}, "user2"))

	// Add Experiments
	suite.location = s2.CellIDFromLatLng(s2.LatLngFromDegrees(1.4093768560366384, 103.79392188731705))
	cell := suite.location.Parent(14)

	dayStart := time.Now().Truncate(24 * time.Hour)
	dayEnd := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)
	hourStart := time.Now().Truncate(time.Hour)
	hourEnd := time.Now().Truncate(time.Hour).Add(time.Hour)

	rawStringSegmenter := []interface{}{"seg-1"}
	rawIntegerSegmenter := []interface{}{1, 2, 3}
	rawIntegerSegmenter2_1 := []interface{}{11, 12}
	rawIntegerSegmenter2_2 := []interface{}{12}
	rawIntegerSegmenter2_3 := []interface{}{12, 13, 14, 15}
	rawBoolSegmenter := []interface{}{true}
	s2Ids, s2IdsEmpty := []interface{}{interface{}(int64(cell.Prev())), interface{}(int64(cell))}, []interface{}{}

	// experiment finished
	addExperiment(newTestXPExperiment(1,
		schema.ExperimentSegment{
			"string_segmenter":    rawStringSegmenter,
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_segmenter_2": rawIntegerSegmenter2_3,
			"s2_ids":              s2Ids,
		}, dayStart, hourStart))

	// experiment will start at next hour (has location segment)
	addExperiment(newTestXPExperiment(1,
		schema.ExperimentSegment{
			"string_segmenter":    rawStringSegmenter,
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_segmenter_2": rawIntegerSegmenter2_2,
			"s2_ids":              s2Ids,
		}, hourEnd, dayEnd))

	// experiment is going and will finish at hour end (has location segment)
	addExperiment(newTestXPExperiment(1,
		schema.ExperimentSegment{
			"string_segmenter":    rawStringSegmenter,
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_segmenter_2": rawIntegerSegmenter2_2,
			"bool_segmenter":      rawBoolSegmenter,
			"s2_ids":              s2Ids,
		}, hourStart, hourEnd))

	// experiment is ongoing and will finish at hour end (no location segment)
	addExperiment(newTestXPExperiment(1,
		schema.ExperimentSegment{
			"string_segmenter":    rawStringSegmenter,
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_segmenter_2": rawIntegerSegmenter2_1,
			"bool_segmenter":      rawBoolSegmenter,
		}, hourStart, hourEnd))

	// experiment is ongoing and will finish at hour end
	// (no string_segmenter, empty list location segment)
	addExperiment(newTestXPExperiment(1,
		schema.ExperimentSegment{
			"integer_segmenter":   rawIntegerSegmenter,
			"integer_segmenter_2": rawIntegerSegmenter2_1,
			"bool_segmenter":      rawBoolSegmenter,
			"s2_ids":              s2IdsEmpty,
		}, hourStart, hourEnd))
}

func (suite *LocalStorageLookupSuite) TestSimpleLookup() {
	loc := s2.CellFromCellID(suite.location).ID()
	locAtLevel14 := int64(loc.Parent(14))
	found := suite.storage.FindExperiments(
		0,
		[]SegmentFilter{
			{Key: "string_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}}}},
			{Key: "integer_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}}}},
			{Key: "integer_segmenter_2", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(11)}}}},
			{Key: "s2_ids", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}}}},
		})

	jsonExperimentMatch := []*ExperimentMatch{
		{
			SegmenterMatches: map[string]Match{
				"string_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}},
				},
				"integer_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}},
				},
				"integer_segmenter_2": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(11)}},
				},
				"s2_ids": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}},
				},
			},
			Experiment: suite.testExperiments[3],
		},
		{
			SegmenterMatches: map[string]Match{
				"string_segmenter": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}},
				},
				"integer_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}},
				},
				"integer_segmenter_2": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(11)}},
				},
				"s2_ids": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}},
				},
			},
			Experiment: suite.testExperiments[4],
		},
	}

	expectedJSON, err := json.Marshal(jsonExperimentMatch)
	require.NoError(suite.T(), err)
	actualJSON, err := json.Marshal(found)
	require.NoError(suite.T(), err)
	assert.JSONEq(suite.T(), string(expectedJSON), string(actualJSON))
}

func (suite *LocalStorageLookupSuite) TestNoExperiments() {
	loc := s2.CellFromCellID(suite.location).ID()
	incorrectLevel := int64(loc)
	found := suite.storage.FindExperiments(
		0,
		[]SegmentFilter{
			{Key: "string_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}}}},
			{Key: "integer_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}}}},
			{Key: "integer_segmenter_2", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(14)}}}},
			{Key: "s2_ids", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: incorrectLevel}}}},
		})
	suite.Require().Equal([]*ExperimentMatch{}, found)
}

func (suite *LocalStorageLookupSuite) TestMultipleExperiments() {
	loc := s2.CellFromCellID(suite.location).ID()
	locAtLevel14 := int64(loc.Parent(14))
	found := suite.storage.FindExperiments(
		0,
		[]SegmentFilter{
			{Key: "string_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}}}},
			{Key: "integer_segmenter", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}}}},
			{Key: "integer_segmenter_2", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(12)}}}},
			{Key: "s2_ids", Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}}}},
		})
	jsonExperimentMatch := []*ExperimentMatch{
		{
			SegmenterMatches: map[string]Match{
				"string_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}},
				},
				"integer_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}},
				},
				"integer_segmenter_2": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(12)}},
				},
				"s2_ids": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}},
				},
			},
			Experiment: suite.testExperiments[2],
		},
		{
			SegmenterMatches: map[string]Match{
				"string_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}},
				},
				"integer_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}},
				},
				"integer_segmenter_2": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(12)}},
				},
				"s2_ids": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}},
				},
			},
			Experiment: suite.testExperiments[3],
		},
		{
			SegmenterMatches: map[string]Match{
				"string_segmenter": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "seg-1"}},
				},
				"integer_segmenter": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(2)}},
				},
				"integer_segmenter_2": {
					MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(12)}},
				},
				"s2_ids": {
					MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: locAtLevel14}},
				},
			},
			Experiment: suite.testExperiments[4],
		},
	}

	expectedJSON, err := json.Marshal(jsonExperimentMatch)
	require.NoError(suite.T(), err)
	actualJSON, err := json.Marshal(found)
	require.NoError(suite.T(), err)
	assert.JSONEq(suite.T(), string(expectedJSON), string(actualJSON))
}

func (suite *LocalStorageLookupSuite) TestInsertAndUpdateExperiment() {
	projectId := int64(1)
	segmentVal := &_segmenters.SegmenterValue_String_{String_: "seg-1"}
	hourStart := time.Now().Truncate(time.Hour)
	hourEnd := time.Now().Truncate(time.Hour).Add(time.Hour)

	experiment := &_pubsub.Experiment{
		Id:        0,
		ProjectId: projectId,
		Status:    _pubsub.Experiment_Active,
		Name:      "new-experiment",
		Segments: map[string]*_segmenters.ListSegmenterValue{"string_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: segmentVal},
		}}},
		Type:       0,
		Interval:   0,
		StartTime:  timestamppb.New(hourStart),
		EndTime:    timestamppb.New(hourEnd),
		Treatments: nil,
		UpdatedAt:  nil,
	}
	suite.storage.InsertExperiment(experiment)
	actual := suite.storage.Experiments[1][0].Experiment.Name

	suite.Require().Equal(experiment.Name, actual)

	hourEnd = time.Now().Truncate(time.Hour).Add(time.Hour * 2)
	experiment = &_pubsub.Experiment{
		Id:        0,
		ProjectId: projectId,
		Status:    _pubsub.Experiment_Active,
		Name:      "new-experiment",
		Segments: map[string]*_segmenters.ListSegmenterValue{"string_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: segmentVal},
		}}},
		Type:       0,
		Interval:   0,
		StartTime:  timestamppb.New(hourStart),
		EndTime:    timestamppb.New(hourEnd),
		Treatments: nil,
		UpdatedAt:  nil,
	}
	suite.storage.UpdateExperiment(experiment)
	found := suite.storage.FindExperimentWithId(1, 0)
	actualName := found.Name
	actualEndTime := found.EndTime

	suite.Require().Equal(found, suite.storage.Experiments[1][0].Experiment)
	suite.Require().Equal(experiment.Name, actualName)
	suite.Require().Equal(experiment.EndTime, actualEndTime)

	// insert inactive experiment
	segmentVal = &_segmenters.SegmenterValue_String_{String_: "seg-1"}
	hourStart = time.Now().Truncate(time.Hour).Add(2 * time.Hour)
	hourEnd = time.Now().Truncate(time.Hour).Add(2 * time.Hour)

	experiment = &_pubsub.Experiment{
		Id:        1,
		ProjectId: projectId,
		Status:    _pubsub.Experiment_Inactive,
		Name:      "new-experiment",
		Segments: map[string]*_segmenters.ListSegmenterValue{"string_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: segmentVal},
		}}},
		Type:       0,
		Interval:   0,
		StartTime:  timestamppb.New(hourStart),
		EndTime:    timestamppb.New(hourEnd),
		Treatments: nil,
		UpdatedAt:  nil,
	}
	suite.storage.InsertExperiment(experiment)
	suite.Require().Equal(1, len(suite.storage.Experiments[ProjectId(projectId)]))

	// update active experiment to inactive
	experiment = &_pubsub.Experiment{
		Id:        0,
		ProjectId: projectId,
		Status:    _pubsub.Experiment_Inactive,
		Name:      "new-experiment",
		Segments: map[string]*_segmenters.ListSegmenterValue{"string_segmenter": {Values: []*_segmenters.SegmenterValue{
			{Value: segmentVal},
		}}},
		Type:       0,
		Interval:   0,
		StartTime:  timestamppb.New(hourStart),
		EndTime:    timestamppb.New(hourEnd),
		Treatments: nil,
		UpdatedAt:  nil,
	}
	suite.storage.UpdateExperiment(experiment)
	suite.Require().Equal(0, len(suite.storage.Experiments[ProjectId(projectId)]))
}

func (suite *LocalStorageLookupSuite) TestGetProjectSettings() {
	projectSettings := suite.storage.FindProjectSettingsWithId(1)

	suite.Require().Equal("user1", projectSettings.Username)
	suite.Require().Equal("passkey", projectSettings.Passkey)
}

func (suite *LocalStorageLookupSuite) TestInsertAndUpdateProjectSettings() {
	projectId := int64(3)
	newSegments := &_pubsub.Segmenters{
		Names: []string{"new-segment", "new-segment2"},
		Variables: map[string]*_pubsub.ExperimentVariables{
			"new-segment":  {Value: []string{"new-segment"}},
			"new-segment2": {Value: []string{"new-segment2"}},
		},
	}
	newProjectSettings := &_pubsub.ProjectSettings{
		ProjectId:  projectId,
		Segmenters: newSegments,
	}
	err := suite.storage.InsertProjectSettings(newProjectSettings)
	suite.NoError(err)
	projectSettings := suite.storage.FindProjectSettingsWithId(ProjectId(projectId))
	suite.Require().Equal(newSegments, projectSettings.Segmenters)

	newSegments = &_pubsub.Segmenters{
		Names: []string{"new-segment3"},
		Variables: map[string]*_pubsub.ExperimentVariables{
			"new-segment3": {Value: []string{"new-segment3"}},
		},
	}
	newProjectSettings = &_pubsub.ProjectSettings{
		ProjectId:  projectId,
		Segmenters: newSegments,
	}
	suite.storage.UpdateProjectSettings(newProjectSettings)
	projectSettings = suite.storage.FindProjectSettingsWithId(ProjectId(projectId))
	suite.Require().Equal(newSegments, projectSettings.Segmenters)
}

func (suite *LocalStorageLookupSuite) TestNewProjectId() {
	projectId := NewProjectId(int64(1))
	expected := uint32(1)

	suite.Require().Equal(expected, projectId)
}

func TestDumpExperiments(t *testing.T) {
	// Set up storage
	s2Ids := []interface{}{int64(3592210796974702592)}
	rawStringSegmenter := []interface{}{"seg-1"}
	rawIntegerSegmenter := []interface{}{int64(1)}
	segmentersType := map[string]schema.SegmenterType{
		"string_segmenter":  "string",
		"integer_segmenter": "integer",
		"s2_ids":            "integer",
	}
	e, err := OpenAPIExperimentSpecToProtobuf(newTestXPExperiment(
		1,
		schema.ExperimentSegment{
			"string_segmenter":  rawStringSegmenter,
			"integer_segmenter": rawIntegerSegmenter,
			"s2_ids":            s2Ids,
		},
		time.Date(2021, 1, 2, 3, 5, 7, 0, time.UTC),
		time.Date(2022, 1, 2, 3, 5, 7, 0, time.UTC),
	), segmentersType)
	require.NoError(t, err)
	storage := LocalStorage{
		Experiments: map[ProjectId][]*ExperimentIndex{
			1: {NewExperimentIndex(e)},
		},
	}
	// Need to hardcode this to testdata/experiments_dump.json ExperimentId
	// because we are randomly generating Id in newTestXPExperiment call
	e.Id = 81

	// Dump experiments
	uuid, _ := uuid.NewUUID()
	filename := fmt.Sprintf("/tmp/%s.json", uuid)
	err = storage.DumpExperiments(filename)
	require.NoError(t, err)
	// Defer file cleanup
	defer func() {
		err := os.Remove(filename)
		if err != nil {
			t.Logf("Error cleaning up file: %s", filename)
		}
	}()

	// Read the JSON content from the result file and the baseline file
	actualBytes, err := tu.ReadFile(filename)
	require.NoError(t, err)
	baselineBytes, err := tu.ReadFile("../testdata/experiments_dump.json")
	require.NoError(t, err)

	// Compare JSON
	assert.Equal(t, true, reflect.DeepEqual(string(baselineBytes), string(actualBytes)))
}

func TestExperimentLookupSuite(t *testing.T) {
	suite.Run(t, new(LocalStorageLookupSuite))
}

func TestExperimentIndexMatchSegment(t *testing.T) {
	stringSetsVal := []interface{}{"test1"}
	intSetsVal := []interface{}{int64(1)}
	realSetsVal := []interface{}{1.0}
	boolSetsVal := []interface{}{true}
	experimentIndex := ExperimentIndex{
		stringSets: map[string]*set.Set{
			"stringType": set.New(stringSetsVal...),
		},
		intSets: map[string]*set.Set{
			"numType": set.New(intSetsVal...),
		},
		realSets: map[string]*set.Set{
			"realType": set.New(realSetsVal...),
		},
		boolSets: map[string]*set.Set{
			"flagType": set.New(boolSetsVal...),
		},
		Experiment: &_pubsub.Experiment{
			Segments: map[string]*_segmenters.ListSegmenterValue{
				"stringType": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_String_{String_: "test1"}},
					},
				},
				"numType": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Integer{Integer: 1}},
					},
				},
				"realType": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}},
					},
				},
				"flagType": {
					Values: []*_segmenters.SegmenterValue{
						{Value: &_segmenters.SegmenterValue_Bool{Bool: true}},
					},
				},
			},
		},
	}
	type args struct {
		segmentName string
		value       []*_segmenters.SegmenterValue
	}
	tests := []struct {
		name string
		args args
		want Match
	}{
		{
			name: "string-type-match",
			args: args{
				segmentName: "stringType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "test1"}}},
			},
			want: Match{MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_String_{String_: "test1"}}},
		},
		{
			name: "string-type-no-match",
			args: args{
				segmentName: "stringType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "test2"}}},
			},
			want: Match{MatchStrengthNone, nil},
		},
		{
			name: "num-type-match",
			args: args{
				segmentName: "numType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(1)}}},
			},
			want: Match{MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Integer{Integer: int64(1)}}},
		},
		{
			name: "real-type-match",
			args: args{
				segmentName: "realType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}}},
			},
			want: Match{MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Real{Real: 1.0}}},
		},
		{
			name: "real-type-no-match",
			args: args{
				segmentName: "realType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Real{Real: 2.0}}},
			},
			want: Match{MatchStrengthNone, nil},
		},
		{
			name: "flag-type-match",
			args: args{
				segmentName: "flagType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: true}}},
			},
			want: Match{MatchStrengthExact, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: true}}},
		},
		{
			name: "flag-type-no-match",
			args: args{
				segmentName: "flagType",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: false}}},
			},
			want: Match{MatchStrengthNone, nil},
		},
		{
			name: "segment-name-dont-exist",
			args: args{
				segmentName: "false",
				value:       []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_Bool{Bool: false}}},
			},
			want: Match{MatchStrengthWeak, &_segmenters.SegmenterValue{Value: &_segmenters.SegmenterValue_Bool{Bool: false}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := experimentIndex.matchSegment(tt.args.segmentName, tt.args.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCustomSegmenter(t *testing.T) {

	projectId := ProjectId(1)
	storage := LocalStorage{
		Experiments:       make(map[ProjectId][]*ExperimentIndex),
		ProjectSegmenters: map[ProjectId]map[string]schema.SegmenterType{projectId: {}},
	}
	assert.Equal(t, 0, len(storage.ProjectSegmenters[projectId]))
	segmenterName := "testseg1"
	segmenterConfig := _segmenters.SegmenterConfiguration{Name: segmenterName, Type: _segmenters.SegmenterValueType_STRING}
	storage.UpdateProjectSegmenters(&segmenterConfig, int64(projectId))
	assert.Equal(t, 1, len(storage.ProjectSegmenters[projectId]))
	segmenterTypeMapping, err := storage.GetSegmentersTypeMapping(1)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToLower(segmenterConfig.Type.String()), string(segmenterTypeMapping[segmenterConfig.Name]))

	segmenterToBeDeleted := _segmenters.SegmenterConfiguration{Name: "testseg2", Type: _segmenters.SegmenterValueType_INTEGER}
	storage.UpdateProjectSegmenters(&segmenterToBeDeleted, int64(projectId))
	assert.Equal(t, 2, len(storage.ProjectSegmenters[projectId]))
	assert.Equal(t, strings.ToLower(segmenterToBeDeleted.Type.String()), string(segmenterTypeMapping[segmenterToBeDeleted.Name]))

	storage.DeleteProjectSegmenters(segmenterToBeDeleted.Name, int64(projectId))
	assert.Equal(t, 1, len(storage.ProjectSegmenters[projectId]))
	assert.Equal(t, strings.ToLower(segmenterConfig.Type.String()), string(segmenterTypeMapping[segmenterConfig.Name]))
	assert.Empty(t, segmenterTypeMapping[segmenterToBeDeleted.Name])

	experiment := newTestXPExperiment(
		int64(projectId),
		schema.ExperimentSegment{
			segmenterName: []interface{}{"stringval"},
		},
		time.Now(),
		time.Now().Add(time.Hour*1))

	e, err := OpenAPIExperimentSpecToProtobuf(experiment, segmenterTypeMapping)
	assert.NoError(t, err)
	storage.Experiments[projectId] = make([]*ExperimentIndex, 0)
	storage.Experiments[projectId] = append(storage.Experiments[projectId], NewExperimentIndex(e))

	experimentmatch := storage.FindExperiments(
		projectId,
		[]SegmentFilter{
			{Key: segmenterName, Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "invalid"}}}}})
	assert.Empty(t, experimentmatch)

	experimentmatch = storage.FindExperiments(
		projectId,
		[]SegmentFilter{
			{Key: segmenterName, Value: []*_segmenters.SegmenterValue{{Value: &_segmenters.SegmenterValue_String_{String_: "stringval"}}}}})
	assert.Equal(t, 1, len(experimentmatch))
}
