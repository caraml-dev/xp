package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/caraml-dev/xp/common/api/schema"
	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	"github.com/caraml-dev/xp/treatment-service/models"
)

type ExperimentService interface {
	// GetExperiment returns experiment after filtering based on required request parameters
	GetExperiment(
		projectId models.ProjectId,
		requestFilter map[string][]*_segmenters.SegmenterValue,
	) ([]models.SegmentFilter, *_pubsub.Experiment, error)
	// DumpExperiments dumps the data in the local storage as a JSON file, in the specified location,
	// and responds with the full file path
	DumpExperiments(directory string) (string, error)
}

type experimentService struct {
	localStorage *models.LocalStorage
}

func NewExperimentService(
	localStorage *models.LocalStorage,
) (ExperimentService, error) {
	svc := &experimentService{
		localStorage: localStorage,
	}

	return svc, nil
}

func (es *experimentService) GetExperiment(
	projectId models.ProjectId,
	requestFilter map[string][]*_segmenters.SegmenterValue,
) ([]models.SegmentFilter, *_pubsub.Experiment, error) {
	// Convert filterParams to Segmenter values
	lookupRequestFilters := es.generateLookupRequest(requestFilter)
	// Retrieve all matching experiments from storage
	matches := es.localStorage.FindExperiments(projectId, lookupRequestFilters)

	projectSettings := es.localStorage.FindProjectSettingsWithId(projectId)
	// Retrieve segmentersTypeMapping that are active with respect to the given project
	segmentersTypeMapping, err := es.localStorage.GetSegmentersTypeMapping(projectId)
	if err != nil {
		return lookupRequestFilters, nil, fmt.Errorf("segmenters cannot be retrieved for projectId: %v",
			projectId)
	}
	// Define filters for resolving experiment based on hierarchy
	type HierarchyFilters func([]*models.ExperimentMatch) []*models.ExperimentMatch
	filters := []HierarchyFilters{
		// Resolve exact vs weak matches, using the inter-segmenter hierarchy
		func(matches []*models.ExperimentMatch) []*models.ExperimentMatch {
			return es.filterByMatchStrength(matches, projectSettings.Segmenters.Names)
		},
		// Resolve granularity of Segmenters
		func(matches []*models.ExperimentMatch) []*models.ExperimentMatch {
			return es.filterByLookupOrder(matches, requestFilter, projectSettings.Segmenters.Names, segmentersTypeMapping)
		},
		// Resolve tiers - at this point, we should ideally only be left with 1 experiment or 2
		// (in different tiers), based on the orthogonality rules enforced by the management service.
		es.filterByTierPriority,
	}

	// While we have more than 1 experiment, progressively apply the filters
	for _, filter := range filters {
		if len(matches) <= 1 {
			break
		}
		matches = filter(matches)
	}

	if len(matches) == 1 {
		return lookupRequestFilters, matches[0].Experiment, nil
	} else if len(matches) > 1 {
		return lookupRequestFilters, nil, errors.New("more than 1 experiment of the same match strength encountered")
	}
	// No experiments matched
	return lookupRequestFilters, nil, nil
}

func (es *experimentService) generateLookupRequest(requestFilter map[string][]*_segmenters.SegmenterValue) []models.SegmentFilter {
	filters := []models.SegmentFilter{}

	for k, v := range requestFilter {
		filters = append(filters, models.SegmentFilter{Key: k, Value: v})
	}

	return filters
}

func (es *experimentService) filterByMatchStrength(
	matches []*models.ExperimentMatch,
	segmenters []string,
) []*models.ExperimentMatch {
	// In the order of the inter-segmenter hierarchy, filter out the weak matches if exact matches exist
	filtered := matches
	for _, segmenter := range segmenters {
		exactMatches := []*models.ExperimentMatch{}
		for _, match := range filtered {
			if segmenterMatch, ok := match.SegmenterMatches[segmenter]; ok && segmenterMatch.Strength == models.MatchStrengthExact {
				exactMatches = append(exactMatches, match)
			}
		}
		if len(exactMatches) > 0 {
			// If we have exact matches for the segmenter, discard the weak ones.
			filtered = exactMatches
		}
	}
	return filtered
}

func (es *experimentService) filterByLookupOrder(
	matches []*models.ExperimentMatch,
	filters map[string][]*_segmenters.SegmenterValue,
	segmenters []string,
	segmenterTypes map[string]schema.SegmenterType,
) []*models.ExperimentMatch {
	// Stop search when we have at least 1 match
	filtered := matches
	for _, segmenter := range segmenters {
		orderedValues := filters[segmenter]
		currentFilteredList := []*models.ExperimentMatch{}
		segmenterType := segmenterTypes[segmenter]
		for _, transformedValue := range orderedValues {
			if len(currentFilteredList) == 0 {
				for _, experiment := range filtered {
					segmenterMatchedValue := experiment.SegmenterMatches[segmenter]
					switch segmenterType {
					case "string":
						if transformedValue.GetString_() == segmenterMatchedValue.Value.GetString_() {
							currentFilteredList = append(currentFilteredList, experiment)
						}
					case "integer":
						if transformedValue.GetInteger() == segmenterMatchedValue.Value.GetInteger() {
							currentFilteredList = append(currentFilteredList, experiment)
						}
					case "real":
						if transformedValue.GetReal() == segmenterMatchedValue.Value.GetReal() {
							currentFilteredList = append(currentFilteredList, experiment)
						}
					case "bool":
						if transformedValue.GetBool() == segmenterMatchedValue.Value.GetBool() {
							currentFilteredList = append(currentFilteredList, experiment)
						}
					}
				}
			}
		}
		// currentFilteredList could be 0 in case of weak matches
		if len(currentFilteredList) > 0 {
			filtered = currentFilteredList
		}
	}

	return filtered
}

func (es *experimentService) filterByTierPriority(matches []*models.ExperimentMatch) []*models.ExperimentMatch {
	overrides := []*models.ExperimentMatch{}
	for _, match := range matches {
		if match.Experiment.Tier == _pubsub.Experiment_Override {
			overrides = append(overrides, match)
		}
	}
	if len(overrides) > 0 {
		return overrides
	}
	// No override experiment(s), return all (i.e., default experiments).
	return matches
}

func (es *experimentService) DumpExperiments(directory string) (string, error) {
	// Create directory if not exists
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return "", err
	}
	// Create file name
	t := time.Now()
	formattedTime := fmt.Sprintf("%d-%02d-%02dT%02d-%02d-%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(),
	)
	filePath := fmt.Sprintf("%s.json", filepath.Join(directory, formattedTime))
	// Log to the file
	return filePath, es.localStorage.DumpExperiments(filePath)
}
