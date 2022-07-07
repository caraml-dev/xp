package appcontext

import (
	"testing"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/xp/management-service/config"
	mw "github.com/gojek/xp/management-service/middleware"
	"github.com/gojek/xp/management-service/services"
	"github.com/gojek/xp/management-service/services/mocks"
)

func TestNewAppContext(t *testing.T) {
	var db *gorm.DB

	cfg := &config.Config{
		AuthorizationConfig: &config.AuthorizationConfig{
			Enabled: true,
			URL:     "http://test-authz-url",
		},
		SegmenterConfig: map[string]interface{}{
			"s2_ids": map[string]interface{}{
				"levels": []int{14},
			},
		},
		MLPConfig: &config.MLPConfig{
			URL: "http://mlp.example.com/api/merlin/v1",
		},
		PubSubConfig: &config.PubSubConfig{
			Project:   "test",
			TopicName: "update",
		},
		ValidationConfig: config.ValidationConfig{
			ValidationUrlTimeoutSeconds: 5,
		},
	}

	// Create members of the App context
	authorizer := &mw.Authorizer{}
	allServices := services.Services{}

	oapiValidator, err := mw.NewOpenAPIValidator(&mw.OpenAPIValidationOptions{
		IgnoreAuthentication: true,
		IgnoreServers:        true,
	})
	require.NoError(t, err)
	segmenterSvc, err := services.NewSegmenterService(&allServices, cfg.SegmenterConfig, db)
	require.NoError(t, err)
	validationService, err := services.NewValidationService(cfg.ValidationConfig)
	require.NoError(t, err)

	pubSubPublisherService, _ := services.NewPubSubPublisherService(cfg.PubSubConfig)

	expHistSvc := services.NewExperimentHistoryService(db)
	expSvc := services.NewExperimentService(&allServices, db)
	projectSettingsSvc := services.NewProjectSettingsService(&allServices, db)
	segmentHistSvc := services.NewSegmentHistoryService(db)
	segmentSvc := services.NewSegmentService(&allServices, db)
	treatmentHistSvc := services.NewTreatmentHistoryService(db)
	treatmentSvc := services.NewTreatmentService(&allServices, db)
	mlpService := &mocks.MLPService{}

	// Patch functions with pointer members, so the result is deterministic
	// Patch the openapi middleware function
	monkey.Patch(
		mw.NewOpenAPIValidator,
		func(*mw.OpenAPIValidationOptions) (*mw.OpenAPIValidator, error) {
			return oapiValidator, nil
		},
	)
	// Patch the Go playground validator service
	monkey.Patch(services.NewValidationService, func(validationConfig config.ValidationConfig) (services.ValidationService, error) {
		return validationService, nil
	})
	// Patch the Segmenter service
	monkey.Patch(services.NewSegmenterService,
		func(
			services *services.Services,
			segmenterConfig map[string]interface{},
			db *gorm.DB,
		) (services.SegmenterService, error) {
			return segmenterSvc, nil
		},
	)
	// Patch PubSub publisher service
	monkey.Patch(services.NewPubSubPublisherService,
		func(pubsubConfig *config.PubSubConfig) (services.PubSubPublisherService, error) {
			return pubSubPublisherService, nil
		},
	)
	// Patch New MLP Service to validate the input and return the mock service object
	monkey.Patch(services.NewMLPService,
		func(mlpBasePath string) (services.MLPService, error) {
			assert.Equal(t, cfg.MLPConfig.URL, mlpBasePath)
			return mlpService, nil
		},
	)

	// Run and validate
	appCtx, err := NewAppContext(db, authorizer, cfg)
	require.NoError(t, err)

	allServices = services.Services{
		ExperimentService:        expSvc,
		ExperimentHistoryService: expHistSvc,
		SegmenterService:         segmenterSvc,
		MLPService:               mlpService,
		ProjectSettingsService:   projectSettingsSvc,
		SegmentService:           segmentSvc,
		SegmentHistoryService:    segmentHistSvc,
		TreatmentService:         treatmentSvc,
		TreatmentHistoryService:  treatmentHistSvc,
		ValidationService:        validationService,
		PubSubPublisherService:   pubSubPublisherService,
	}
	monkey.Patch(services.NewServices,
		func(
			service services.ExperimentService,
			historyService services.ExperimentHistoryService,
			segmenterService services.SegmenterService,
			mlpService services.MLPService,
			settingsService services.ProjectSettingsService,
			segmentSvc services.SegmentService,
			segmentHistoryService services.SegmentHistoryService,
			treatmentService services.TreatmentService,
			treatmentHistoryService services.TreatmentHistoryService,
			validationService services.ValidationService,
			publisherService services.PubSubPublisherService,
		) services.Services {
			return allServices
		},
	)

	assert.Equal(t, &AppContext{
		Authorizer:       authorizer,
		OpenAPIValidator: oapiValidator,
		Services:         allServices,
	}, appCtx)
}
