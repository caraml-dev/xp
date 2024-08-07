// Package treatment provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.8.1 DO NOT EDIT.
package treatment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	externalRef0 "github.com/caraml-dev/xp/common/api/schema"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/pkg/errors"
)

// FetchTreatmentBadRequest defines model for FetchTreatmentBadRequest.
type FetchTreatmentBadRequest externalRef0.Error

// FetchTreatmentSuccess defines model for FetchTreatmentSuccess.
type FetchTreatmentSuccess struct {
	Data *externalRef0.SelectedTreatment `json:"data,omitempty"`
}

// InternalServerError defines model for InternalServerError.
type InternalServerError externalRef0.Error

// FetchTreatmentRequestBody defines model for FetchTreatmentRequestBody.
type FetchTreatmentRequestBody struct {
	AdditionalProperties map[string]interface{} `json:"-"`
}

// FetchTreatmentParams defines parameters for FetchTreatment.
type FetchTreatmentParams struct {
	PassKey string `json:"pass-key"`
}

// FetchTreatmentJSONRequestBody defines body for FetchTreatment for application/json ContentType.
type FetchTreatmentJSONRequestBody FetchTreatmentRequestBody

// Getter for additional properties for FetchTreatmentRequestBody. Returns the specified
// element and whether it was found
func (a FetchTreatmentRequestBody) Get(fieldName string) (value interface{}, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for FetchTreatmentRequestBody
func (a *FetchTreatmentRequestBody) Set(fieldName string, value interface{}) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]interface{})
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for FetchTreatmentRequestBody to handle AdditionalProperties
func (a *FetchTreatmentRequestBody) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]interface{})
		for fieldName, fieldBuf := range object {
			var fieldVal interface{}
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for FetchTreatmentRequestBody to handle AdditionalProperties
func (a FetchTreatmentRequestBody) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// FetchTreatment request  with any body
	FetchTreatmentWithBody(ctx context.Context, projectId int64, params *FetchTreatmentParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	FetchTreatment(ctx context.Context, projectId int64, params *FetchTreatmentParams, body FetchTreatmentJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) FetchTreatmentWithBody(ctx context.Context, projectId int64, params *FetchTreatmentParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewFetchTreatmentRequestWithBody(c.Server, projectId, params, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) FetchTreatment(ctx context.Context, projectId int64, params *FetchTreatmentParams, body FetchTreatmentJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewFetchTreatmentRequest(c.Server, projectId, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewFetchTreatmentRequest calls the generic FetchTreatment builder with application/json body
func NewFetchTreatmentRequest(server string, projectId int64, params *FetchTreatmentParams, body FetchTreatmentJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewFetchTreatmentRequestWithBody(server, projectId, params, "application/json", bodyReader)
}

// NewFetchTreatmentRequestWithBody generates requests for FetchTreatment with any type of body
func NewFetchTreatmentRequestWithBody(server string, projectId int64, params *FetchTreatmentParams, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "project_id", runtime.ParamLocationPath, projectId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/projects/%s/fetch-treatment", pathParam0)
	if operationPath[0] == '/' {
		operationPath = operationPath[1:]
	}
	operationURL := url.URL{
		Path: operationPath,
	}

	queryURL := serverURL.ResolveReference(&operationURL)

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	var headerParam0 string

	headerParam0, err = runtime.StyleParamWithLocation("simple", false, "pass-key", runtime.ParamLocationHeader, params.PassKey)
	if err != nil {
		return nil, err
	}

	req.Header.Set("pass-key", headerParam0)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// FetchTreatment request  with any body
	FetchTreatmentWithBodyWithResponse(ctx context.Context, projectId int64, params *FetchTreatmentParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*FetchTreatmentResponse, error)

	FetchTreatmentWithResponse(ctx context.Context, projectId int64, params *FetchTreatmentParams, body FetchTreatmentJSONRequestBody, reqEditors ...RequestEditorFn) (*FetchTreatmentResponse, error)
}

type FetchTreatmentResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Data *externalRef0.SelectedTreatment `json:"data,omitempty"`
	}
	JSON400 *externalRef0.Error
	JSON500 *externalRef0.Error
}

// Status returns HTTPResponse.Status
func (r FetchTreatmentResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r FetchTreatmentResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// FetchTreatmentWithBodyWithResponse request with arbitrary body returning *FetchTreatmentResponse
func (c *ClientWithResponses) FetchTreatmentWithBodyWithResponse(ctx context.Context, projectId int64, params *FetchTreatmentParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*FetchTreatmentResponse, error) {
	rsp, err := c.FetchTreatmentWithBody(ctx, projectId, params, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseFetchTreatmentResponse(rsp)
}

func (c *ClientWithResponses) FetchTreatmentWithResponse(ctx context.Context, projectId int64, params *FetchTreatmentParams, body FetchTreatmentJSONRequestBody, reqEditors ...RequestEditorFn) (*FetchTreatmentResponse, error) {
	rsp, err := c.FetchTreatment(ctx, projectId, params, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseFetchTreatmentResponse(rsp)
}

// ParseFetchTreatmentResponse parses an HTTP response from a FetchTreatmentWithResponse call
func ParseFetchTreatmentResponse(rsp *http.Response) (*FetchTreatmentResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	if err != nil {
		return nil, err
	}

	response := &FetchTreatmentResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Data *externalRef0.SelectedTreatment `json:"data,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest externalRef0.Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest externalRef0.Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}
