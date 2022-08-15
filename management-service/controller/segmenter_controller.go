package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/caraml-dev/xp/common/api/schema"
	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/models"
	"github.com/caraml-dev/xp/management-service/services"
)

type SegmenterController struct {
	*appcontext.AppContext
}

func NewSegmenterController(ctx *appcontext.AppContext) *SegmenterController {
	return &SegmenterController{ctx}
}

func (s SegmenterController) ListSegmenters(w http.ResponseWriter, r *http.Request, projectId int64, params api.ListSegmentersParams) {
	// Perform validation checks on the projectId given; note that instead of simply returning the error directly to
	// the user if the project corresponding to the project id is not found in the db, we are returning a list of global
	// segmenters. This temporary behaviour is implemented in order to allow the create settings UI page to run
	// correctly after the removal of the '/segmenters' endpoint
	if err := s.validateProjectId(projectId); err != nil {
		segmenters, err := s.Services.SegmenterService.ListGlobalSegmenters()
		if err != nil {
			WriteErrorResponse(w, err)
			return
		}
		Ok(w, &segmenters)
		return
	}

	var scope *services.SegmenterScope
	if params.Scope != nil {
		val, ok := services.SegmenterScopeMap[string(*params.Scope)]
		if !ok {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "scope passed is not a string representing segmenter scope"))
			return
		}
		scope = &val
	}
	var status *services.SegmenterStatus
	if params.Status != nil {
		val, ok := services.SegmenterStatusMap[string(*params.Status)]
		if !ok {
			WriteErrorResponse(w, errors.Newf(errors.BadInput, "status passed is not a string representing segmenter status"))
			return
		}
		status = &val
	}

	queryParams := services.ListSegmentersParams{
		Scope:  scope,
		Status: status,
		Search: params.Search,
	}
	allSegmenters, err := s.Services.SegmenterService.ListSegmenters(projectId, queryParams)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, &allSegmenters)
}

func (s SegmenterController) GetSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	// Perform validation checks on the projectId given
	if err := s.validateProjectId(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Check all segmenters if a segmenter with a matching name exists
	segmenter, err := s.Services.SegmenterService.GetSegmenter(projectId, name)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, segmenter)
}

func (s SegmenterController) CreateSegmenter(w http.ResponseWriter, r *http.Request, projectId int64) {
	// Parse request body
	customSegmenterData := api.CreateSegmenterRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&customSegmenterData); err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}
	// Perform validation checks on the projectId given
	if err := s.validateProjectId(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Create custom segmenter
	customSegmenter, err := s.Services.SegmenterService.CreateCustomSegmenter(
		projectId,
		toCreateCustomSegmenterBody(customSegmenterData),
	)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, customSegmenter.ToApiSchema())
}

func (s SegmenterController) UpdateSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	// Parse request body
	customSegmenterData := api.UpdateSegmenterRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&customSegmenterData)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}

	// Update custom segmenter
	customSegmenter, err := s.Services.SegmenterService.UpdateCustomSegmenter(
		projectId,
		name,
		toUpdateCustomSegmenterBody(customSegmenterData),
	)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}

	Ok(w, customSegmenter.ToApiSchema())
}

func (s SegmenterController) DeleteSegmenter(w http.ResponseWriter, r *http.Request, projectId int64, name string) {
	// Perform validation checks on the projectId given
	if err := s.validateProjectId(projectId); err != nil {
		WriteErrorResponse(w, err)
		return
	}
	// Delete selected custom segmenter
	err := s.Services.SegmenterService.DeleteCustomSegmenter(projectId, name)
	if err != nil {
		WriteErrorResponse(w, err)
		return
	}
	resp := map[string]string{"name": name}

	Ok(w, resp)
}

func toCreateCustomSegmenterBody(body api.CreateSegmenterRequestBody) services.CreateCustomSegmenterRequestBody {
	return services.CreateCustomSegmenterRequestBody{
		Name:        body.Name,
		Type:        strings.ToUpper(string(body.Type)),
		Options:     parseApiOptions(body.Options),
		MultiValued: body.MultiValued,
		Constraints: parseApiConstraints(body.Constraints),
		Required:    body.Required,
		Description: body.Description,
	}
}

func toUpdateCustomSegmenterBody(body api.UpdateSegmenterRequestBody) services.UpdateCustomSegmenterRequestBody {
	return services.UpdateCustomSegmenterRequestBody{
		Options:     parseApiOptions(body.Options),
		MultiValued: body.MultiValued,
		Constraints: parseApiConstraints(body.Constraints),
		Required:    body.Required,
		Description: body.Description,
	}
}

func parseApiConstraints(apiConstraints *[]schema.Constraint) *models.Constraints {
	if apiConstraints == nil {
		return nil
	}
	var constraints models.Constraints
	for _, constraint := range *apiConstraints {
		constraints = append(
			constraints,
			models.Constraint{
				PreRequisites: parseApiConstraintPreRequisites(constraint.PreRequisites),
				AllowedValues: parseApiConstraintSegmenterValues(constraint.AllowedValues),
				Options:       parseApiOptions(constraint.Options),
			})
	}
	return &constraints
}

func parseApiOptions(apiOptions *schema.SegmenterOptions) *models.Options {
	if apiOptions == nil || apiOptions.AdditionalProperties == nil {
		return nil
	}
	options := make(models.Options)
	for key, value := range apiOptions.AdditionalProperties {
		options[key] = value
	}
	return &options
}

func parseApiConstraintPreRequisites(apiPreRequisites []schema.PreRequisite) []models.PreRequisite {
	var preRequisites []models.PreRequisite
	for _, preRequisite := range apiPreRequisites {
		segmenterValues := parseApiConstraintSegmenterValues(preRequisite.SegmenterValues)
		preRequisites = append(
			preRequisites,
			models.PreRequisite{
				SegmenterName:   preRequisite.SegmenterName,
				SegmenterValues: segmenterValues,
			})
	}
	return preRequisites
}

func parseApiConstraintSegmenterValues(
	apiSegmenterValues []schema.SegmenterValues,
) []interface{} {
	var segmenterValues []interface{}
	for _, segmenterValue := range apiSegmenterValues {
		segmenterValues = append(segmenterValues, segmenterValue)
	}
	return segmenterValues
}

func (s SegmenterController) validateProjectId(projectId int64) error {
	// Check if the projectId is valid
	if _, err := s.Services.MLPService.GetProject(projectId); err != nil {
		return err
	}
	// Check if the projectId has been set up
	_, err := s.Services.ProjectSettingsService.GetProjectSettings(projectId)
	if err != nil {
		return errors.Wrapf(err, "Settings for project_id %d cannot be retrieved", projectId)
	}
	return nil
}
