package services_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gojek/xp/management-service/config"
	"github.com/gojek/xp/management-service/models"
	"github.com/gojek/xp/management-service/services"
)

// testHTTPServerAddr is the address of the test HTTP server assumed to be running and
// serving the validation endpoints required by the validation service during testing.
var testHTTPServerAddr = "127.0.0.1:9000"

// successEndpoint is the endpoint of the test HTTP server address to serve validation requests with a successful
// response
var successEndpoint = "/validate-with-success"

// failureEndpoint is the endpoint of the test HTTP server address to serve validation requests with an unsuccessful
// response
var failureEndpoint = "/validate-with-failure"

// invalidEndpoint is the endpoint of the test HTTP server address that is invalid
var invalidEndpoint = "/invalid-endpoint"

type ValidationServiceTestSuite struct {
	suite.Suite
	services.ValidationService

	stopServer func()
}

func (s *ValidationServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ValidationServiceTestSuite")
	svc, err := services.NewValidationService(config.ValidationConfig{ValidationUrlTimeoutSeconds: 5})
	if err != nil {
		s.Suite.T().Fatalf("Could not set up test data: %v", err)
	}

	// Set up Test HTTP Server
	stopServer, err := startTestHTTPServer(testHTTPServerAddr)
	if err != nil {
		s.Suite.T().Fatalf("Failed to start test http server: " + err.Error())
	}
	s.stopServer = stopServer

	s.ValidationService = svc
}

// This test server can be used to test sending requests to custom validation endpoints.
func startTestHTTPServer(addr string) (stopServer func(), err error) {
	handler := http.NewServeMux()
	handler.HandleFunc(successEndpoint, validationSuccessHandler)
	handler.HandleFunc(failureEndpoint, validationFailureHandler)
	server := httptest.NewUnstartedServer(handler)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	server.Listener = listener
	server.Start()

	return func() { server.Close() }, nil
}

func validationSuccessHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(`{"done"}`))
}

func validationFailureHandler(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusInternalServerError)
	_, _ = rw.Write([]byte(`{"done"}`))
}

func (s *ValidationServiceTestSuite) TearDownSuite() {
	s.Suite.T().Log("Cleaning up ValidationServiceTestSuite")
	s.stopServer()
}

func TestValidationService(t *testing.T) {
	suite.Run(t, new(ValidationServiceTestSuite))
}

func (s *ValidationServiceTestSuite) TestCreateProjectSettings() {
	testValidationUrl := "https://test-me.io"
	tests := map[string]struct {
		data      services.CreateProjectSettingsRequestBody
		errString string
	}{
		"failure | blank randomization key": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "abc",
				RandomizationKey: " ",
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.RandomizationKey' Error:Field validation for 'RandomizationKey' failed on the 'notBlank' tag",
		},
		"failure | no name": {
			data: services.CreateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.Username' Error:Field validation for 'Username' failed on the 'required' tag",
		},
		"failure | blank treatment schema rule name": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      " ",
							Predicate: "predicate_1",
						},
					},
				},
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.TreatmentSchema.Rules[0].Name' " +
				"Error:Field validation for 'Name' failed on the 'notBlank' tag",
		},
		"failure | blank treatment schema rule predicate": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: " ",
						},
					},
				},
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.TreatmentSchema.Rules[0].Predicate' " +
				"Error:Field validation for 'Predicate' failed on the 'notBlank' tag",
		},
		"failure | non-unique treatment schema rule names": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_1",
							Predicate: "predicate_2",
						},
					},
				},
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.TreatmentSchema.Rules' " +
				"Error:Field validation for 'Rules' failed on the 'unique' tag",
		},
		"failure | invalid treatment schema predicate": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "valid-rule",
							Predicate: "{{- ( .field1) -}}",
						},
						{
							Name:      "invalid-rule",
							Predicate: "{{{{{",
						},
					},
				},
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.TreatmentSchema' Error:Field " +
				"validation for 'TreatmentSchema' failed on the 'invalid-predicate: template: " +
				":1: unexpected \"{\" in command' tag",
		},
		"failure | invalid validation url": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				ValidationUrl:    new(string),
			},
			errString: "Key: 'CreateProjectSettingsRequestBody.ValidationUrl' " +
				"Error:Field validation for 'ValidationUrl' failed on the 'url' tag",
		},

		"success | no segmenters": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
			},
		},
		"success | valid treatment schema rules": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_2",
							Predicate: "predicate_2",
						},
					},
				},
			},
		},
		"success | valid validation url": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
				ValidationUrl:    &testValidationUrl,
			},
		},
		"success": {
			data: services.CreateProjectSettingsRequestBody{
				Username:         "name",
				RandomizationKey: "rkey",
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestUpdateProjectSettings() {
	testValidationUrl := "https://test-me.io"
	enableClustering := false
	tests := map[string]struct {
		data      services.UpdateProjectSettingsRequestBody
		errString string
	}{
		"failure | blank randomization key": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: " ",
			},
			errString: "Key: 'UpdateProjectSettingsRequestBody.RandomizationKey' " +
				"Error:Field validation for 'RandomizationKey' failed on the 'notBlank' tag",
		},
		"failure | blank treatment schema rule name": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      " ",
							Predicate: "predicate_1",
						},
					},
				},
			},
			errString: "Key: 'UpdateProjectSettingsRequestBody.TreatmentSchema.Rules[0].Name' " +
				"Error:Field validation for 'Name' failed on the 'notBlank' tag",
		},
		"failure | blank treatment schema rule predicate": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: " ",
						},
					},
				},
			},
			errString: "Key: 'UpdateProjectSettingsRequestBody.TreatmentSchema.Rules[0].Predicate' " +
				"Error:Field validation for 'Predicate' failed on the 'notBlank' tag",
		},
		"failure | non-unique treatment schema rule names": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_1",
							Predicate: "predicate_2",
						},
					},
				},
			},
			errString: "Key: 'UpdateProjectSettingsRequestBody.TreatmentSchema.Rules' " +
				"Error:Field validation for 'Rules' failed on the 'unique' tag",
		},
		"failure | invalid validation url": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				ValidationUrl:    new(string),
			},
			errString: "Key: 'UpdateProjectSettingsRequestBody.ValidationUrl' " +
				"Error:Field validation for 'ValidationUrl' failed on the 'url' tag",
		},
		"success | no segmenters": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
			},
		},
		"success | valid treatment schema rules": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				TreatmentSchema: &models.TreatmentSchema{
					Rules: []models.Rule{
						{
							Name:      "rule_1",
							Predicate: "predicate_1",
						},
						{
							Name:      "rule_2",
							Predicate: "predicate_2",
						},
					},
				},
			},
		},
		"success | valid validation url": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey: "rkey",
				ValidationUrl:    &testValidationUrl,
			},
		},
		"success": {
			data: services.UpdateProjectSettingsRequestBody{
				RandomizationKey:     "rkey",
				EnableS2idClustering: &enableClustering,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestCreateExperimentRequestParameters() {
	description := "desc"
	negativeInterval := int32(-1)
	interval := int32(10)
	traffic0 := int32(0)
	traffic50 := int32(50)
	traffic100 := int32(100)
	updatedBy := "testuser"
	blankUpdatedBy := " "
	name1234 := "1234"
	name1234Repeated := "1234"
	name4567 := "4567"
	nameValid := "abcd"
	nameInvalid := "abc abc "
	experimentSegment := models.ExperimentSegmentRaw{}
	tests := map[string]struct {
		data      services.CreateExperimentRequestBody
		errString string
	}{
		"failure | interval set a/b": {
			data: services.CreateExperimentRequestBody{
				Name:        nameValid,
				Description: &description,
				EndTime:     time.Now().Add(time.Hour),
				Interval:    &interval,
				Segment:     experimentSegment,
				StartTime:   time.Now().Add(time.Minute),
				Status:      models.ExperimentStatusInactive,
				Treatments:  []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:        models.ExperimentTierDefault,
				Type:        models.ExperimentTypeAB,
				UpdatedBy:   &updatedBy,
			},
			errString: "Key: 'CreateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-unset-ab-experiment' tag",
		},
		"failure | interval unset switchback": {
			data: services.CreateExperimentRequestBody{
				Name:        nameValid,
				Description: &description,
				EndTime:     time.Now().Add(time.Hour),
				Segment:     experimentSegment,
				StartTime:   time.Now().Add(time.Minute),
				Status:      models.ExperimentStatusInactive,
				Treatments:  []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:        models.ExperimentTierDefault,
				Type:        models.ExperimentTypeSwitchback,
				UpdatedBy:   &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Interval' Error:Field validation for 'Interval' ",
				"failed on the 'interval-set-switchback-experiment' tag"}, ""),
		},
		"failure | required fields": {
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.EndTime' Error:Field validation for 'EndTime' failed on the 'required' tag",
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				"Key: 'CreateExperimentRequestBody.StartTime' Error:Field validation for 'StartTime' failed on the 'required' tag",
				"Key: 'CreateExperimentRequestBody.Status' Error:Field validation for 'Status' failed on the 'required' tag",
				"Key: 'CreateExperimentRequestBody.Tier' Error:Field validation for 'Tier' failed on the 'required' tag",
				"Key: 'CreateExperimentRequestBody.Type' Error:Field validation for 'Type' failed on the 'required' tag",
				strings.Join([]string{
					"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
					"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
					"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
				}, " "),
				"Key: 'CreateExperimentRequestBody.StartTime' Error:Field validation for 'StartTime' failed on the 'start-time-in-future' tag",
				"Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'notBlank' tag",
			}, "\n"),
		},
		"failure | not blank fields": {
			data: services.CreateExperimentRequestBody{
				Name:      " ",
				Tier:      models.ExperimentTierDefault,
				Type:      models.ExperimentTypeSwitchback,
				Status:    models.ExperimentStatusActive,
				StartTime: time.Now().Add(time.Minute),
				EndTime:   time.Now().Add(10 * time.Minute),
				UpdatedBy: &blankUpdatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the 'notBlank' tag",
				strings.Join([]string{
					"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
					"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
					"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
				}, " "),
				"Key: 'CreateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-set-switchback-experiment' tag",
				"Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'notBlank' tag",
			}, "\n"),
		},
		"failure | traffic not 100": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic50}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
			errString: "Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-sum-100' tag",
		},
		"failure | Switchback 0 traffic treatment": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}, {Name: name4567, Traffic: &traffic0}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: "Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-is-0' tag",
		},
		"failure | A/B 0 traffic treatment": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}, {Name: name4567, Traffic: &traffic0}},
				Tier:       models.ExperimentTierOverride,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-unset-ab-experiment' tag\n",
				"Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-is-0' tag",
			}, ""),
		},
		"failure | end time before start time": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Minute),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Hour),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
			errString: "Key: 'CreateExperimentRequestBody.EndTime' Error:Field validation for 'EndTime' failed on the 'gtfield' tag",
		},
		"failure | negative interval": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &negativeInterval,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Interval' Error:Field validation for 'Interval' ",
				"failed on the 'interval-set-switchback-experiment' tag"}, ""),
		},
		"failure | incorrect regex for experiment name": {
			data: services.CreateExperimentRequestBody{
				Name:       "abc abc ",
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | incorrect regex for treatment name": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: nameInvalid, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | non-unique treatment name": {
			data: services.CreateExperimentRequestBody{
				Name:      nameValid,
				EndTime:   time.Now().Add(time.Hour),
				Interval:  &interval,
				Segment:   experimentSegment,
				StartTime: time.Now().Add(time.Minute),
				Status:    models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{
					{Name: name1234, Traffic: &traffic50},
					{Name: name1234Repeated, Traffic: &traffic50},
				},
				Tier:      models.ExperimentTierDefault,
				Type:      models.ExperimentTypeSwitchback,
				UpdatedBy: &updatedBy,
			},
			errString: "Key: 'CreateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'unique' tag",
		},
		"success | ab": {
			data: services.CreateExperimentRequestBody{
				Name:       nameValid,
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
		},
		"success | switchback": {
			data: services.CreateExperimentRequestBody{
				Name:      nameValid,
				EndTime:   time.Now().Add(time.Hour),
				Interval:  &interval,
				Segment:   experimentSegment,
				StartTime: time.Now().Add(time.Minute),
				Status:    models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{
					{Name: name1234, Traffic: &traffic50},
					{Name: name4567, Traffic: &traffic50},
				},
				Tier:      models.ExperimentTierDefault,
				Type:      models.ExperimentTypeSwitchback,
				UpdatedBy: &updatedBy,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestUpdateExperimentRequestParameters() {
	description := "desc"
	negativeInterval := int32(-1)
	interval := int32(10)
	traffic0 := int32(0)
	traffic50 := int32(50)
	traffic100 := int32(100)
	updatedBy := "testuser"
	blankUpdatedBy := " "
	name1234 := "1234"
	name1234Repeated := "1234"
	name4567 := "4567"
	nameInvalid := "abc abc "
	experimentSegment := models.ExperimentSegmentRaw{}
	tests := map[string]struct {
		data      services.UpdateExperimentRequestBody
		errString string
	}{
		"failure | interval set a/b": {
			data: services.UpdateExperimentRequestBody{
				Description: &description,
				EndTime:     time.Now().Add(time.Hour),
				Interval:    &interval,
				Segment:     experimentSegment,
				StartTime:   time.Now().Add(time.Minute),
				Status:      models.ExperimentStatusInactive,
				Treatments:  []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:        models.ExperimentTierDefault,
				Type:        models.ExperimentTypeAB,
				UpdatedBy:   &updatedBy,
			},
			errString: "Key: 'UpdateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-unset-ab-experiment' tag",
		},
		"failure | interval unset switchback": {
			data: services.UpdateExperimentRequestBody{
				Description: &description,
				EndTime:     time.Now().Add(time.Hour),
				Segment:     experimentSegment,
				StartTime:   time.Now().Add(time.Minute),
				Status:      models.ExperimentStatusInactive,
				Treatments:  []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:        models.ExperimentTierOverride,
				Type:        models.ExperimentTypeSwitchback,
				UpdatedBy:   &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.Interval' Error:Field validation for 'Interval' ",
				"failed on the 'interval-set-switchback-experiment' tag"}, ""),
		},
		"failure | required fields": {
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.EndTime' Error:Field validation for 'EndTime' failed on the 'required' tag",
				"Key: 'UpdateExperimentRequestBody.StartTime' Error:Field validation for 'StartTime' failed on the 'required' tag",
				"Key: 'UpdateExperimentRequestBody.Status' Error:Field validation for 'Status' failed on the 'required' tag",
				"Key: 'UpdateExperimentRequestBody.Tier' Error:Field validation for 'Tier' failed on the 'required' tag",
				"Key: 'UpdateExperimentRequestBody.Type' Error:Field validation for 'Type' failed on the 'required' tag",
				"Key: 'UpdateExperimentRequestBody.StartTime' Error:Field validation for 'StartTime' failed on the 'start-time-in-future' tag",
				"Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'notBlank' tag",
			}, "\n"),
		},
		"failure | not blank fields": {
			data: services.UpdateExperimentRequestBody{
				Tier:      models.ExperimentTierDefault,
				Type:      models.ExperimentTypeSwitchback,
				Status:    models.ExperimentStatusActive,
				StartTime: time.Now().Add(time.Minute),
				EndTime:   time.Now().Add(10 * time.Minute),
				UpdatedBy: &blankUpdatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-set-switchback-experiment' tag",
				"Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'notBlank' tag",
			}, "\n"),
		},
		"failure | traffic not 100": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic50}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
			errString: "Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-sum-100' tag",
		},
		"failure | Switchback 0 traffic treatment": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}, {Name: name4567, Traffic: &traffic0}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: "Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-is-0' tag",
		},
		"failure | A/B 0 traffic treatment": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}, {Name: name4567, Traffic: &traffic0}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.Interval' Error:Field validation for 'Interval' failed on the 'interval-unset-ab-experiment' tag\n",
				"Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'traffic-is-0' tag",
			}, ""),
		},
		"failure | end time before start time": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Minute),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Hour),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
			errString: "Key: 'UpdateExperimentRequestBody.EndTime' Error:Field validation for 'EndTime' failed on the 'gtfield' tag",
		},
		"failure | negative interval": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &negativeInterval,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.Interval' Error:Field validation for 'Interval' ",
				"failed on the 'interval-set-switchback-experiment' tag"}, ""),
		},
		"failure | incorrect regex for treatment name": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: nameInvalid, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
				Interval:   &interval,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | non-unique treatment name": {
			data: services.UpdateExperimentRequestBody{
				EndTime:   time.Now().Add(time.Hour),
				Interval:  &interval,
				Segment:   experimentSegment,
				StartTime: time.Now().Add(time.Minute),
				Status:    models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{
					{Name: name1234, Traffic: &traffic50},
					{Name: name1234Repeated, Traffic: &traffic50},
				},
				Tier:      models.ExperimentTierDefault,
				Type:      models.ExperimentTypeSwitchback,
				UpdatedBy: &updatedBy,
			},
			errString: "Key: 'UpdateExperimentRequestBody.Treatments' Error:Field validation for 'Treatments' failed on the 'unique' tag",
		},
		"success | ab default": {
			data: services.UpdateExperimentRequestBody{
				EndTime:    time.Now().Add(time.Hour),
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234, Traffic: &traffic100}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeAB,
				UpdatedBy:  &updatedBy,
			},
		},
		"success | switchback override": {
			data: services.UpdateExperimentRequestBody{
				EndTime:   time.Now().Add(time.Hour),
				Interval:  &interval,
				Segment:   experimentSegment,
				StartTime: time.Now().Add(time.Minute),
				Status:    models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{
					{Name: name1234, Traffic: &traffic50},
					{Name: name4567, Traffic: &traffic50},
				},
				Tier:      models.ExperimentTierOverride,
				Type:      models.ExperimentTypeSwitchback,
				UpdatedBy: &updatedBy,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestCreateTreatmentRequestParameters() {
	updatedBy := "testuser"
	blankUpdatedBy := " "
	nameInvalid := "abc abc "
	tests := map[string]struct {
		data      services.CreateTreatmentRequestBody
		errString string
	}{
		"failure | required fields": {
			errString: strings.Join([]string{
				"Key: 'CreateTreatmentRequestBody.Config' Error:Field validation for 'Config' failed on the 'required' tag",
				"Key: 'CreateTreatmentRequestBody.Name' Error:Field validation for 'Name' failed on the 'required' tag",
				strings.Join([]string{
					"Key: 'CreateTreatmentRequestBody.Name' Error:Field validation for 'Name' failed on the",
					"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
					"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
				}, " "),
			}, "\n"),
		},
		"failure | not blank fields": {
			data: services.CreateTreatmentRequestBody{
				Name:      " ",
				Config:    map[string]interface{}{},
				UpdatedBy: &blankUpdatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateTreatmentRequestBody.Name' Error:Field validation for 'Name' failed on the 'notBlank' tag",
				strings.Join([]string{
					"Key: 'CreateTreatmentRequestBody.Name' Error:Field validation for 'Name' failed on the",
					"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
					"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
				}, " "),
			}, "\n"),
		},
		"failure | invalid name": {
			data: services.CreateTreatmentRequestBody{
				Name:      nameInvalid,
				Config:    map[string]interface{}{"key": "value"},
				UpdatedBy: &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateTreatmentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"success": {
			data: services.CreateTreatmentRequestBody{
				Name:      "test-treatment",
				Config:    map[string]interface{}{"team": "ds"},
				UpdatedBy: &updatedBy,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestUpdateTreatmentRequestParameters() {
	updatedBy := "testuser"
	blankUpdatedBy := " "
	tests := map[string]struct {
		data      services.UpdateTreatmentRequestBody
		errString string
	}{
		"failure | required fields": {
			errString: strings.Join([]string{
				"Key: 'UpdateTreatmentRequestBody.Config' Error:Field validation for 'Config' failed on the 'required' tag",
			}, "\n"),
		},
		"failure | not blank fields": {
			data: services.UpdateTreatmentRequestBody{
				Config:    map[string]interface{}{},
				UpdatedBy: &blankUpdatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'UpdateTreatmentRequestBody.Config' Error:Field validation for 'Config' failed on the 'notBlank' tag",
			}, "\n"),
		},
		"success": {
			data: services.UpdateTreatmentRequestBody{
				Config:    map[string]interface{}{"team": "ds"},
				UpdatedBy: &updatedBy,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestCreateExperimentNameRegex() {
	interval := int32(10)
	updatedBy := "testuser"
	name1234 := "1234"
	experimentSegment := models.ExperimentSegmentRaw{}
	tests := map[string]struct {
		data      services.CreateExperimentRequestBody
		errString string
	}{
		"failure | Trailing space before": {
			data: services.CreateExperimentRequestBody{
				Name:       " abcd",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | Trailing space after": {
			data: services.CreateExperimentRequestBody{
				Name:       "abcd ",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | Less than 4 characters": {
			data: services.CreateExperimentRequestBody{
				Name:       "abc",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | More than 64 characters": {
			data: services.CreateExperimentRequestBody{
				Name:       "abcdefghijklmnopqrstuvwxyz abcdefghijklmnopqrstuvwxyz abcdefghijklmnopqrstuvwxyz",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | Invalid trailing character": {
			data: services.CreateExperimentRequestBody{
				Name:       "abc@",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
			errString: strings.Join([]string{
				"Key: 'CreateExperimentRequestBody.Name' Error:Field validation for 'Name' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"success | switchback with valid symbol": {
			data: services.CreateExperimentRequestBody{
				Name:       "aBc -_()#$%&:.",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
		},
		"success | switchback": {
			data: services.CreateExperimentRequestBody{
				Name:       "aBc -_()#$%&:.d",
				EndTime:    time.Now().Add(time.Hour),
				Interval:   &interval,
				Segment:    experimentSegment,
				StartTime:  time.Now().Add(time.Minute),
				Status:     models.ExperimentStatusInactive,
				Treatments: []models.ExperimentTreatment{{Name: name1234}},
				Tier:       models.ExperimentTierDefault,
				Type:       models.ExperimentTypeSwitchback,
				UpdatedBy:  &updatedBy,
			},
		},
	}

	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.Validate(data.data)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestValidateDataWithValidationUrl() {
	treatment := models.Treatment{
		Configuration: models.TreatmentConfig{
			"field1": "abc",
			"field2": "def",
			"field3": map[string]interface{}{
				"field4": 0.1,
			},
		},
	}
	data := treatment.Configuration
	context := services.ValidationContext{}
	validationSuccessUrl := "http://" + testHTTPServerAddr + successEndpoint
	validationFailureUrl := "http://" + testHTTPServerAddr + failureEndpoint
	invalidUrl := "http://" + testHTTPServerAddr + invalidEndpoint

	tests := map[string]struct {
		operation     services.OperationType
		entityType    services.EntityType
		data          map[string]interface{}
		context       services.ValidationContext
		validationUrl *string
		errString     string
	}{
		"failure | error when sending request to invalid endpoint": {
			operation:     services.OperationTypeCreate,
			entityType:    services.EntityTypeTreatment,
			data:          data,
			context:       context,
			validationUrl: &invalidUrl,
			errString:     "Error validating data with validation URL: 404 Not Found",
		},
		"failure | validation failure from url provided": {
			operation:     services.OperationTypeCreate,
			entityType:    services.EntityTypeTreatment,
			data:          data,
			context:       context,
			validationUrl: &validationFailureUrl,
			errString:     "Error validating data with validation URL: 500 Internal Server Error",
		},
		"success | no validation url provided": {
			operation:     services.OperationTypeCreate,
			entityType:    services.EntityTypeTreatment,
			data:          data,
			context:       context,
			validationUrl: nil,
		},
		"success": {
			operation:     services.OperationTypeCreate,
			entityType:    services.EntityTypeTreatment,
			data:          data,
			context:       context,
			validationUrl: &validationSuccessUrl,
		},
	}

	for name, test := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := s.ValidationService.ValidateEntityWithExternalUrl(
				test.operation,
				test.entityType,
				test.data,
				test.context,
				test.validationUrl,
			)
			if test.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, test.errString)
			}
		})
	}
}

func (s *ValidationServiceTestSuite) TestValidateRulePredicate() {
	tests := map[string]struct {
		predicate string
		errString string
	}{
		"failure | invalid predicate": {
			predicate: "{{ myFunction }}",
			errString: "template: :1: function \"myFunction\" not defined",
		},
		"success | valid predicate": {
			predicate: "{{- ( .field1) -}}",
		},
	}
	for name, data := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			err := services.CheckRulePredicate(data.predicate)
			if data.errString == "" {
				s.Suite.Require().NoError(err)
			} else {
				s.Suite.Assert().EqualError(err, data.errString)
			}
		})
	}
}
