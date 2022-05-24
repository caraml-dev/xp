package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
)

const (
	resourceSegmenters = "segmenters"
	resourceValidate   = "validate"
)

type Authorizer struct {
	authEnforcer enforcer.Enforcer
}

// NewAuthorizer creates a new authorization middleware using the given auth enforcer
func NewAuthorizer(enforcer enforcer.Enforcer) (*Authorizer, error) {
	// Set up XP API specific policies
	err := upsertSegmentersListAllPolicy(enforcer)
	if err != nil {
		return nil, err
	}

	err = upsertValidationPolicy(enforcer)
	if err != nil {
		return nil, err
	}

	return &Authorizer{authEnforcer: enforcer}, nil
}

var methodMapping = map[string]string{
	http.MethodGet:     enforcer.ActionRead,
	http.MethodHead:    enforcer.ActionRead,
	http.MethodPost:    enforcer.ActionCreate,
	http.MethodPut:     enforcer.ActionUpdate,
	http.MethodPatch:   enforcer.ActionUpdate,
	http.MethodDelete:  enforcer.ActionDelete,
	http.MethodConnect: enforcer.ActionRead,
	http.MethodOptions: enforcer.ActionRead,
	http.MethodTrace:   enforcer.ActionRead,
}

func (a *Authorizer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resource := getResourceFromPath(r.URL.Path)
		action := getActionFromMethod(r.Method)
		user := r.Header.Get("User-Email")

		allowed, err := a.authEnforcer.Enforce(user, resource, action)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}
		if !*allowed {
			jsonError(w,
				fmt.Sprintf("%s is not authorized to execute %s on %s", user, action, resource),
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getResourceFromPath(path string) string {
	return strings.Replace(strings.TrimPrefix(path, "/"), "/", ":", -1)
}

func getActionFromMethod(method string) string {
	return methodMapping[method]
}

func upsertSegmentersListAllPolicy(authEnforcer enforcer.Enforcer) error {

	// Upsert policy
	policyName := fmt.Sprintf("allow-all-list-%s", resourceSegmenters)
	_, err := authEnforcer.UpsertPolicy(
		policyName,
		[]string{},
		[]string{"**"},
		[]string{resourceSegmenters},
		[]string{enforcer.ActionRead},
	)
	return err
}

func upsertValidationPolicy(authEnforcer enforcer.Enforcer) error {

	// Upsert policy
	policyName := "validation-policy"
	_, err := authEnforcer.UpsertPolicy(
		policyName,
		[]string{},
		[]string{"**"},
		[]string{resourceValidate},
		[]string{enforcer.ActionCreate},
	)
	return err
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if len(msg) > 0 {
		errJSON, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{msg})

		_, _ = w.Write(errJSON)
	}
}
