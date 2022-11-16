package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	managementClient "github.com/caraml-dev/xp/clients/management"
	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/golang-collections/collections/set"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const ExpGoogleApplicationCredentials = "EXP_GOOGLE_APPLICATION_CREDENTIALS"

var GoogleOAuthScope = "https://www.googleapis.com/auth/userinfo.email"

type ProjectId = uint32
type StringSet = map[string]interface{}
type IntSet = map[int64]interface{}
type RealSet = map[float64]interface{}

type MatchStrength string

const (
	MatchStrengthExact MatchStrength = "exact"
	MatchStrengthWeak  MatchStrength = "weak"
	MatchStrengthNone  MatchStrength = "none"
)

type ExperimentMatch struct {
	Experiment       *pubsub.Experiment
	SegmenterMatches map[string]Match
}

type ProjectSettingsStorage interface {
	FindProjectSettingsWithId(projectId ProjectId) *pubsub.ProjectSettings
}

type ExperimentStorage interface {
	FindExperiments(projectId ProjectId, filters []SegmentFilter) []*ExperimentMatch
	FindExperimentWithId(projectId ProjectId, experimentId int64) *pubsub.Experiment
	InsertExperiment(experiment *pubsub.Experiment)
	DeactivateExperiment(projectId ProjectId, experimentId int64) error
	// DumpExperiments is a helper method for a Debug API
	DumpExperiments(filepath string) error
}

type LocalStorage struct {
	sync.RWMutex
	Experiments          map[ProjectId][]*ExperimentIndex
	ProjectSettings      []*pubsub.ProjectSettings
	managementClient     *managementClient.ClientWithResponses
	subscribedProjectIds []ProjectId
	Segmenters           map[string]schema.SegmenterType
	ProjectSegmenters    map[ProjectId]map[string]schema.SegmenterType
}

type Match struct {
	Strength MatchStrength
	Value    *_segmenters.SegmenterValue
}

type SegmentFilter struct {
	Key   string
	Value []*_segmenters.SegmenterValue
}

type ExperimentIndex struct {
	stringSets map[string]*set.Set
	intSets    map[string]*set.Set
	realSets   map[string]*set.Set

	StartTime time.Time
	EndTime   time.Time

	Experiment *pubsub.Experiment
}

// ExperimentIndexLog captures the critical information from the ExperimentIndex,
// in a concise manner, for logging.
type ExperimentIndexLog struct {
	StringSets map[string][]interface{}
	IntSets    map[string][]interface{}
	RealSets   map[string][]interface{}

	StartTime time.Time
	EndTime   time.Time

	ExperimentId int64
	Status       string
	Tier         string
}

// MarshalJSON is a custom marshal function that only includes the critical info
func (i *ExperimentIndex) MarshalJSON() ([]byte, error) {
	// Convert value maps to lists
	stringSets := map[string][]interface{}{}
	for k, v := range i.stringSets {
		values := []interface{}{}
		v.Do(func(item interface{}) {
			values = append(values, item)
		})
		stringSets[k] = values
	}
	intSets := map[string][]interface{}{}
	for k, v := range i.intSets {
		values := []interface{}{}
		v.Do(func(item interface{}) {
			values = append(values, item)
		})
		intSets[k] = values
	}
	realSets := map[string][]interface{}{}
	for k, v := range i.realSets {
		values := []interface{}{}
		v.Do(func(item interface{}) {
			values = append(values, item)
		})
		realSets[k] = values
	}

	idx := ExperimentIndexLog{
		StringSets: stringSets,
		IntSets:    intSets,
		RealSets:   realSets,
		StartTime:  i.StartTime,
		EndTime:    i.EndTime,
	}

	// Store experiment info
	if i.Experiment != nil {
		idx.ExperimentId = i.Experiment.Id
		idx.Status = i.Experiment.Status.String()
		idx.Tier = i.Experiment.Tier.String()
	}

	return json.Marshal(idx)
}

func (i *ExperimentIndex) matchFlagSetSegment(segmentName string, value bool) MatchStrength {
	if _, exists := i.Experiment.Segments[segmentName]; !exists ||
		len(i.Experiment.Segments[segmentName].GetValues()) == 0 {
		// Optional segmenter
		return MatchStrengthWeak
	}

	for _, val := range i.Experiment.Segments[segmentName].GetValues() {
		if val.GetBool() == value {
			return MatchStrengthExact
		}
	}

	return MatchStrengthNone
}

func (i *ExperimentIndex) matchStringSetSegment(segmentName string, value string) MatchStrength {
	set, exists := i.stringSets[segmentName]
	if !exists || set.Len() == 0 {
		// Optional segmenter
		return MatchStrengthWeak
	}

	if set.Has(value) {
		return MatchStrengthExact
	}
	return MatchStrengthNone
}

func (i *ExperimentIndex) matchIntSetSegment(segmentName string, value int64) MatchStrength {
	set, exists := i.intSets[segmentName]
	if !exists || set.Len() == 0 {
		// Optional segmenter
		return MatchStrengthWeak
	}

	if set.Has(value) {
		return MatchStrengthExact
	}
	return MatchStrengthNone
}

func (i *ExperimentIndex) matchRealSetSegment(segmentName string, value float64) MatchStrength {
	set, exists := i.realSets[segmentName]
	if !exists || set.Len() == 0 {
		// Optional segmenter
		return MatchStrengthWeak
	}

	if set.Has(value) {
		return MatchStrengthExact
	}
	return MatchStrengthNone
}

func (i *ExperimentIndex) matchSegment(segmentName string, values []*_segmenters.SegmenterValue) Match {
	if len(values) == 0 {
		// We can either have an optional match on the experiment or none.
		if _, exists := i.Experiment.Segments[segmentName]; !exists ||
			len(i.Experiment.Segments[segmentName].GetValues()) == 0 {
			return Match{Strength: MatchStrengthWeak, Value: nil}
		}
	}

	matchStrength := MatchStrengthNone
	for _, v := range values {
		switch v.Value.(type) {
		case *_segmenters.SegmenterValue_Bool:
			matchStrength = i.matchFlagSetSegment(segmentName, v.GetBool())
		case *_segmenters.SegmenterValue_String_:
			matchStrength = i.matchStringSetSegment(segmentName, v.GetString_())
		case *_segmenters.SegmenterValue_Integer:
			matchStrength = i.matchIntSetSegment(segmentName, v.GetInteger())
		case *_segmenters.SegmenterValue_Real:
			matchStrength = i.matchRealSetSegment(segmentName, v.GetReal())
		}
		if matchStrength != MatchStrengthNone {
			return Match{Strength: matchStrength, Value: v}
		}
	}

	return Match{Strength: matchStrength, Value: nil}
}

func (i *ExperimentIndex) isActive() bool {
	if i.Experiment.Status != pubsub.Experiment_Active {
		return false
	}

	return (i.StartTime.Before(time.Now()) || i.StartTime.Equal(time.Now())) && i.EndTime.After(time.Now())
}

func (s *LocalStorage) InsertProjectSettings(projectSettings *pubsub.ProjectSettings) error {
	s.Lock()
	defer s.Unlock()

	// check that settings with the same Id doesn't exist
	for _, existingSettings := range s.ProjectSettings {
		if existingSettings.GetProjectId() == projectSettings.GetProjectId() {
			return nil
		}
	}

	s.ProjectSettings = append(s.ProjectSettings, projectSettings)
	// Update project segmenters on creation
	if err := s.initProjectSegmenters([]*pubsub.ProjectSettings{projectSettings}); err != nil {
		return err
	}
	return nil
}

func (s *LocalStorage) UpdateProjectSettings(updatedProjectSettings *pubsub.ProjectSettings) {
	s.Lock()
	defer s.Unlock()

	for index, settings := range s.ProjectSettings {
		if updatedProjectSettings.ProjectId == settings.ProjectId {
			s.ProjectSettings[index] = updatedProjectSettings
		}
	}
}

func (s *LocalStorage) FindProjectSettingsWithId(projectId ProjectId) *pubsub.ProjectSettings {
	s.RLock()
	defer s.RUnlock()

	if !ContainsProjectId(s.subscribedProjectIds, projectId) {
		return nil
	}

	for _, settings := range s.ProjectSettings {
		if ProjectId(settings.ProjectId) == projectId {
			return settings
		}
	}

	// In case new project was just created and we are subscribed to its ID
	// we'll try to retrieve it from management service
	projectSettingsResponse, err := s.managementClient.GetProjectSettingsWithResponse(
		context.Background(), int64(projectId))
	if err != nil {
		return nil
	}

	project := OpenAPIProjectSettingsSpecToProtobuf(projectSettingsResponse.JSON200.Data)
	s.ProjectSettings = append(s.ProjectSettings, project)
	return project
}

func (s *LocalStorage) GetSegmentersTypeMapping(projectId ProjectId) (map[string]schema.SegmenterType, error) {
	s.RLock()
	defer s.RUnlock()

	if segmenters, ok := s.ProjectSegmenters[projectId]; ok {
		return segmenters, nil
	} else {
		return nil, errors.New("project segmenter not found for project id: " + fmt.Sprint(projectId))
	}
}

func (s *LocalStorage) FindExperiments(projectId ProjectId, filters []SegmentFilter) []*ExperimentMatch {
	experiments := s.Experiments[projectId]
	s.RLock()
	defer s.RUnlock()

	var matched = make([]*ExperimentMatch, 0)

	for _, item := range experiments {
		if !item.isActive() {
			continue
		}

		// Match all segmenters
		matchStrengths := map[string]Match{}
		match := true
		for _, filter := range filters {
			matchStrengths[filter.Key] = item.matchSegment(filter.Key, filter.Value)
			if matchStrengths[filter.Key].Strength == MatchStrengthNone {
				match = false
				break
			}
		}

		if match {
			matched = append(matched, &ExperimentMatch{
				Experiment:       item.Experiment,
				SegmenterMatches: matchStrengths,
			})
		}
	}

	return matched
}

func (s *LocalStorage) FindExperimentWithId(projectId ProjectId, experimentId int64) *pubsub.Experiment {
	currentExperiments, settingsExist := s.Experiments[projectId]

	s.RLock()
	defer s.RUnlock()
	if !settingsExist {
		return nil
	}

	for _, existingIndex := range currentExperiments {
		if existingIndex.Experiment.Id == experimentId {
			return existingIndex.Experiment
		}
	}

	return nil
}

func NewExperimentIndex(experiment *pubsub.Experiment) *ExperimentIndex {
	stringSets := make(map[string]*set.Set)
	intSets := make(map[string]*set.Set)
	realSets := make(map[string]*set.Set)

	for key, segment := range experiment.Segments {
		for _, val := range segment.Values {
			switch val.Value.(type) {
			case *_segmenters.SegmenterValue_String_:
				_, ok := stringSets[key]
				if !ok {
					stringSets[key] = set.New()
				}
				stringSets[key].Insert(val.GetString_())
			case *_segmenters.SegmenterValue_Integer:
				_, ok := intSets[key]
				if !ok {
					intSets[key] = set.New()
				}
				intSets[key].Insert(val.GetInteger())
			case *_segmenters.SegmenterValue_Real:
				_, ok := realSets[key]
				if !ok {
					realSets[key] = set.New()
				}
				realSets[key].Insert(val.GetReal())
			}
		}
	}

	return &ExperimentIndex{
		Experiment: experiment,
		stringSets: stringSets,
		intSets:    intSets,
		realSets:   realSets,
		StartTime:  time.Unix(experiment.StartTime.Seconds, 0).UTC(),
		EndTime:    time.Unix(experiment.EndTime.Seconds, 0).UTC(),
	}
}

func (s *LocalStorage) InsertExperiment(experiment *pubsub.Experiment) {
	projectId := ProjectId(experiment.ProjectId)
	s.Lock()
	defer s.Unlock()

	// do not add inactive experiment in local storage
	if experiment.Status == pubsub.Experiment_Inactive {
		return
	}

	// check that experiment with the same Id doesn't exist
	for _, existingIndex := range s.Experiments[projectId] {
		if existingIndex.Experiment.Id == experiment.Id {
			return
		}
	}

	newIndex := NewExperimentIndex(experiment)
	s.Experiments[projectId] = append(s.Experiments[projectId], newIndex)
}

func (s *LocalStorage) UpdateExperiment(experiment *pubsub.Experiment) {
	projectId := ProjectId(experiment.ProjectId)
	s.Lock()
	defer s.Unlock()
	newIndex := NewExperimentIndex(experiment)

	experimentIndexes := s.Experiments[projectId]
	for idx, experimentIndex := range experimentIndexes {
		if experimentIndex.Experiment.Id == experiment.Id {
			if experimentIndex.Experiment.Status == pubsub.Experiment_Active && experiment.Status == pubsub.Experiment_Inactive {
				// do not keep inactive experiment in local storage
				indexToRemove := set.New()
				indexToRemove.Insert(idx)
				updatedExperimentIndexes := removeExperiment(experimentIndexes, *indexToRemove)
				s.Experiments[projectId] = updatedExperimentIndexes
			} else {
				experimentIndexes[idx] = newIndex
			}
			return
		}
	}
	// previously disabled experiment is enabled again
	s.Experiments[projectId] = append(s.Experiments[projectId], newIndex)
}

// DumpExperiments is used to dump the experiment from the local cache into the
// given file, as JSON. Useful for debugging.
func (s *LocalStorage) DumpExperiments(filepath string) error {
	s.RLock()
	defer s.RUnlock()

	file, err := json.MarshalIndent(s.Experiments, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, file, 0644)
}

func (s *LocalStorage) init() error {
	s.Lock()
	defer s.Unlock()
	log.Println("retrieving projects...")
	listProjectsResponse, err := s.managementClient.ListProjectsWithResponse(context.Background())
	if err != nil {
		return err
	}
	if listProjectsResponse.StatusCode() != http.StatusOK {
		errMessage := ""
		if listProjectsResponse.JSON500 != nil {
			errMessage = listProjectsResponse.JSON500.Message
		}
		return fmt.Errorf("error retrieving projectSettings from xp (%d): %s", listProjectsResponse.StatusCode(), errMessage)
	}
	subscribedProjectSettings := make([]*pubsub.ProjectSettings, 0)
	log.Println("retrieving project settings...")
	for _, project := range listProjectsResponse.JSON200.Data {
		// Get the full settings details of each individual project
		projectSettingsResponse, err := s.managementClient.GetProjectSettingsWithResponse(context.Background(), project.Id)
		if err != nil {
			return err
		}
		if ContainsProjectId(s.subscribedProjectIds, ProjectId(project.Id)) {
			subscribedProjectSettings = append(
				subscribedProjectSettings,
				OpenAPIProjectSettingsSpecToProtobuf(projectSettingsResponse.JSON200.Data),
			)
		}
	}
	if len(s.subscribedProjectIds) > 0 && len(subscribedProjectSettings) != len(s.subscribedProjectIds) {
		return errors.New("not all subscribed project ids are found")
	}
	s.ProjectSettings = subscribedProjectSettings

	err = s.initProjectSegmenters(subscribedProjectSettings)
	if err != nil {
		return err
	}

	err = s.initExperiments(subscribedProjectSettings)
	if err != nil {
		return err
	}

	return nil
}

func NewLocalStorage(projectIds []ProjectId, xpServer string, authzEnabled bool) (*LocalStorage, error) {
	// Set up Request Modifiers
	clientOptions := []managementClient.ClientOption{}
	if authzEnabled {
		var googleClient *http.Client
		// Init Google client for Authz. XP Server must be behind MLP Auth proxy, for authorization to work.
		if filepath := os.Getenv(ExpGoogleApplicationCredentials); filepath != "" {
			data, err := os.ReadFile(filepath)
			if err != nil {
				return nil, err
			}
			creds, err := google.CredentialsFromJSON(context.Background(), data, GoogleOAuthScope)
			if err != nil {
				return nil, err
			}
			googleClient = oauth2.NewClient(context.Background(), creds.TokenSource)
		} else {
			var err error
			googleClient, err = google.DefaultClient(context.Background(), GoogleOAuthScope)
			if err != nil {
				return nil, err
			}
		}

		clientOptions = append(
			clientOptions,
			managementClient.WithHTTPClient(googleClient),
		)
	}
	xpClient, err := managementClient.NewClientWithResponses(xpServer, clientOptions...)
	if err != nil {
		return nil, err
	}
	segmenterCache := make(map[ProjectId]map[string]schema.SegmenterType)
	s := LocalStorage{managementClient: xpClient, subscribedProjectIds: projectIds, ProjectSegmenters: segmenterCache}
	err = s.init()

	return &s, err
}

func (s *LocalStorage) initExperiments(subscribedProjectSettings []*pubsub.ProjectSettings) error {
	log.Println("retrieving project experiments...")
	index := make(map[ProjectId][]*ExperimentIndex)
	for _, projectSettings := range subscribedProjectSettings {
		log.Printf("retrieving experiments for %d", projectSettings.ProjectId)
		projectId := ProjectId(projectSettings.ProjectId)
		startTime := time.Now()
		endTime := time.Now().Add(855360 * time.Hour)
		activeStatus := schema.ExperimentStatusActive

		segmentersType := s.ProjectSegmenters[projectId]
		resp, err := s.managementClient.ListExperimentsWithResponse(
			context.TODO(),
			projectSettings.ProjectId,
			&managementClient.ListExperimentsParams{StartTime: &startTime, EndTime: &endTime, Status: &activeStatus},
		)
		if err != nil {
			return err
		}

		if resp.StatusCode() == 200 {
			projectExperiments := resp.JSON200.Data
			index[projectId] = make([]*ExperimentIndex, 0)
			index, err = flattenProjectExperiments(projectId, index, projectExperiments, segmentersType)
			if err != nil {
				return err
			}

			var pages int
			if resp.JSON200.Paging != nil {
				pages = int(resp.JSON200.Paging.Pages)
			}
			for i := 2; i <= pages; i++ {
				page := int32(i)
				resp, err := s.managementClient.ListExperimentsWithResponse(
					context.TODO(),
					projectSettings.ProjectId,
					&managementClient.ListExperimentsParams{Page: &page, StartTime: &startTime, EndTime: &endTime, Status: &activeStatus},
				)
				if err != nil {
					return err
				}
				if resp.StatusCode() == 200 {
					projectExperiments := resp.JSON200.Data
					index, err = flattenProjectExperiments(projectId, index, projectExperiments, segmentersType)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	s.Experiments = index

	return nil
}

func (s *LocalStorage) initProjectSegmenters(settings []*pubsub.ProjectSettings) error {
	for _, projectSettings := range settings {
		log.Printf("retrieving project segmenters for %d", projectSettings.ProjectId)
		segmentersResp, err := s.managementClient.ListSegmentersWithResponse(
			context.TODO(),
			projectSettings.ProjectId,
			&managementClient.ListSegmentersParams{},
		)
		if err != nil {
			return err
		}
		segmenters := map[string]schema.SegmenterType{}
		for _, v := range segmentersResp.JSON200.Data {
			segmenters[v.Name] = schema.SegmenterType(strings.ToLower(string(v.Type)))
		}
		s.ProjectSegmenters[ProjectId(projectSettings.ProjectId)] = segmenters
	}
	return nil
}

func (s *LocalStorage) UpdateProjectSegmenters(segmenter *_segmenters.SegmenterConfiguration, projectId int64) {
	s.Lock()
	defer s.Unlock()
	s.ProjectSegmenters[ProjectId(projectId)][segmenter.Name] = schema.SegmenterType(strings.ToLower(segmenter.Type.String()))
}

func (s *LocalStorage) DeleteProjectSegmenters(segmenterName string, projectId int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.ProjectSegmenters[ProjectId(projectId)], segmenterName)
}

func NewProjectId(id int64) ProjectId {
	return ProjectId(uint32(id))
}

func flattenProjectExperiments(
	projectId ProjectId,
	projectExperiments map[ProjectId][]*ExperimentIndex,
	experiments []schema.Experiment,
	segmentersType map[string]schema.SegmenterType,
) (map[ProjectId][]*ExperimentIndex, error) {
	for _, projectExperiment := range experiments {
		protoRecord, err := OpenAPIExperimentSpecToProtobuf(projectExperiment, segmentersType)
		if err != nil {
			return projectExperiments, err
		}
		projectExperiments[projectId] = append(
			projectExperiments[projectId],
			NewExperimentIndex(protoRecord),
		)
	}

	return projectExperiments, nil
}

func removeExperiment(experimentIndexes []*ExperimentIndex, indicesToRemove set.Set) []*ExperimentIndex {
	newIndices := []*ExperimentIndex{}
	for idx, experimentIndex := range experimentIndexes {
		if !indicesToRemove.Has(idx) {
			newIndices = append(newIndices, experimentIndex)
		}
	}

	return newIndices
}
