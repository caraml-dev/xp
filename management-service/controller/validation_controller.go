package controller

import (
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/xp/management-service/api"
	"github.com/caraml-dev/xp/management-service/appcontext"
	"github.com/caraml-dev/xp/management-service/errors"
	"github.com/caraml-dev/xp/management-service/services"
)

type ValidationController struct {
	*appcontext.AppContext
}

func NewValidationController(ctx *appcontext.AppContext) *ValidationController {
	return &ValidationController{ctx}
}

func (v ValidationController) ValidateEntity(w http.ResponseWriter, r *http.Request) {
	validationRequest := api.ValidateEntityRequestBody{}
	err := json.NewDecoder(r.Body).Decode(&validationRequest)
	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
		return
	}

	// Return a 400 error if both/neither ValidationUrl and TreatmentSchema are provided
	if (validationRequest.ValidationUrl != nil && validationRequest.TreatmentSchema != nil) ||
		(validationRequest.ValidationUrl == nil && validationRequest.TreatmentSchema == nil) {
		WriteErrorResponse(w, errors.Newf(errors.BadInput,
			"Both/neither the validation url and treatment schema are set"))
		return
	}

	// Perform custom validation with the payload specified in validationRequest.Data if the external validation url is
	// set; if treatment schema is set, perform validation with treatment schema specified in validationRequest.Data
	if validationRequest.ValidationUrl != nil {
		var reqBody []byte
		reqBody, err = json.Marshal(validationRequest.Data)
		if err != nil {
			WriteErrorResponse(w, errors.Newf(errors.BadInput,
				"Error marshalling the validation data: %v", err.Error()))
			return
		}
		err = v.Services.ValidationService.ValidateWithExternalUrl(reqBody, validationRequest.ValidationUrl)
	} else if validationRequest.TreatmentSchema != nil {
		treatmentSchema := parseTreatmentSchema(validationRequest.TreatmentSchema)

		for _, rule := range treatmentSchema.Rules {
			err = services.CheckRulePredicate(rule.Predicate)
			if err != nil {
				WriteErrorResponse(w, errors.Newf(errors.BadInput, err.Error()))
				return
			}
		}

		err = services.ValidateTreatmentConfigWithTreatmentSchema(
			validationRequest.Data,
			parseTreatmentSchema(validationRequest.TreatmentSchema),
		)
	}

	if err != nil {
		WriteErrorResponse(w, errors.Newf(errors.Unknown, err.Error()))
		return
	}

}
