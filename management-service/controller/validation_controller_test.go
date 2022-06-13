package controller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/turing-experiments/management-service/appcontext"
	"github.com/gojek/turing-experiments/management-service/errors"
	"github.com/gojek/turing-experiments/management-service/services"
	"github.com/gojek/turing-experiments/management-service/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var successValidationUrl = "https://validation-success.io"

var failureValidationUrl = "https://validation-failure.io"

type ValidationControllerTestSuite struct {
	suite.Suite
	ctrl                        *ValidationController
	expectedErrorResponseFormat string
}

func (s *ValidationControllerTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up ValidationControllerTestSuite")

	// Configure expected responses and errors
	s.expectedErrorResponseFormat = `{"code":"%[1]v", "error":%[2]v, "message":%[2]v}`

	// Create mock validation service and set up with test responses
	validationSvc := &mocks.ValidationService{}

	validationSvc.
		On("ValidateWithExternalUrl",
			mock.Anything,
			&successValidationUrl).
		Return(nil)

	validationSvc.
		On("ValidateWithExternalUrl",
			mock.Anything,
			&failureValidationUrl).
		Return(errors.Newf(errors.BadInput, "Error validating data with validation URL: 500 Internal Server Error"))

	// Create test controller
	s.ctrl = &ValidationController{
		AppContext: &appcontext.AppContext{
			Services: services.Services{
				ValidationService: validationSvc,
			},
		},
	}
}

func TestValidationController(t *testing.T) {
	suite.Run(t, new(ValidationControllerTestSuite))
}

func (p *ValidationControllerTestSuite) TestValidateEntity() {
	t := p.Suite.T()

	tests := []struct {
		name        string
		requestBody string
		expected    string
	}{
		{
			name:        "failure | no treatment schema or validation url specified",
			requestBody: `{}`,
			expected: `{"code":"400", "error":"Both/neither the validation url and treatment schema are set", 
"message":"Both/neither the validation url and treatment schema are set"}`,
		},
		{
			name:        "failure | bad input",
			requestBody: `{badinput}`,
			expected: `{"code":"400","error":"invalid character 'b' looking for beginning of object key string",
"message":"invalid character 'b' looking for beginning of object key string"}`,
		},
		{
			name: "failure | bad treatment schema input",
			requestBody: `{"data":{"field1":"abc","field2":"def","field3":{"field4":0.1}},
"treatment_schema":{"rules":[{"name":"test-rule","predicate":"{{- (eq .field1 \\\"abc\\\") -}}"}]}}`,
			expected: `{"code":"400","error":"template: :1: unexpected \"\\\\\" in operand",
"message":"template: :1: unexpected \"\\\\\" in operand"}`,
		},
		{
			name:        "failure | validation with external url but error was returned",
			requestBody: fmt.Sprintf(`{"validation_url": "%s"}`, failureValidationUrl),
			expected: `{"code":"500", "error":"Error validating data with validation URL: 500 Internal Server Error", 
"message":"Error validating data with validation URL: 500 Internal Server Error"}`,
		},
		{
			name:        "success | validation with external url and passes",
			requestBody: fmt.Sprintf(`{"validation_url": "%s"}`, successValidationUrl),
			expected:    ``,
		},
		{
			name: "failure | validation with treatment schema but error was returned",
			requestBody: `{"data":{"field1":"abc","field2":"def","field3":{"field4":0.1}},
"treatment_schema":{"rules":[{"name":"test-rule","predicate":"{{- (eq .field1 \"def\") -}}"}]}}`,
			expected: `{"code":"500", "error":"Go template rule test-rule returns false", 
"message":"Go template rule test-rule returns false"}`,
		},
		{
			name: "success | validation with treatment schema and passes",
			requestBody: `{"data":{"field1":"abc","field2":"def","field3":{"field4":0.1}},
"treatment_schema":{"rules":[{"name":"test-rule","predicate":"{{- (eq .field1 \"abc\") -}}"}]}}`,
			expected: ``,
		},
	}
	// Run tests
	for _, data := range tests {
		t.Run(data.name, func(t *testing.T) {
			// Make test requests
			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(data.requestBody)))
			p.Suite.Require().NoError(err)
			w := httptest.NewRecorder()
			// Test error response
			p.ctrl.ValidateEntity(w, req)
			resp := w.Result()
			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}
			body, err := io.ReadAll(resp.Body)
			p.Suite.Require().NoError(err)
			if data.expected != `` {
				p.Suite.Assert().JSONEq(data.expected, string(body))
			} else {
				p.Suite.Assert().Equal(len(body), 0)
			}
		})
	}
}
