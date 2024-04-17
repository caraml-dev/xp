package appcontext

import (
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	common_mq_config "github.com/caraml-dev/xp/common/messagequeue"
	"github.com/caraml-dev/xp/management-service/config"
	mw "github.com/caraml-dev/xp/management-service/middleware"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/messagequeue"
	"github.com/caraml-dev/xp/management-service/services/mocks"
)

func TestNewAppContext(t *testing.T) {
	var db *gorm.DB

	cfg := &config.Config{
		SegmenterConfig: map[string]interface{}{
			"s2_ids": map[string]interface{}{
				"levels": []int{14},
			},
		},
		MLPConfig: &config.MLPConfig{
			URL: "http://mlp.example.com/api/merlin/v1",
		},
		MessageQueueConfig: &common_mq_config.MessageQueueConfig{
			PubSubConfig: &common_mq_config.PubSubConfig{
				Project:   "test",
				TopicName: "update",
			},
		},
		ValidationConfig: config.ValidationConfig{
			ValidationUrlTimeoutSeconds: 5,
		},
	}

	// Create members of the App context
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

	messageQueueService, _ := messagequeue.NewMessageQueueService(*cfg.MessageQueueConfig)

	expHistSvc := services.NewExperimentHistoryService(db)
	expSvc := services.NewExperimentService(&allServices, db)
	projectSettingsSvc := services.NewProjectSettingsService(&allServices, db)
	segmentHistSvc := services.NewSegmentHistoryService(db)
	segmentSvc := services.NewSegmentService(&allServices, db)
	treatmentHistSvc := services.NewTreatmentHistoryService(db)
	treatmentSvc := services.NewTreatmentService(&allServices, db)
	mlpService := &mocks.MLPService{}
	configurationSvc := services.NewConfigurationService(cfg)

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
	// Patch MessageQueue service
	monkey.Patch(messagequeue.NewMessageQueueService,
		func(messageQueueConfig common_mq_config.MessageQueueConfig) (messagequeue.MessageQueueService, error) {
			return messageQueueService, nil
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
	appCtx, err := NewAppContext(db, cfg)
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
		MessageQueueService:      messageQueueService,
		ConfigurationService:     configurationSvc,
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
			messageQueueService messagequeue.MessageQueueService,
			configurationService services.ConfigurationService,
		) services.Services {
			return allServices
		},
	)

	assert.Equal(t, &AppContext{
		OpenAPIValidator: oapiValidator,
		Services:         allServices,
	}, appCtx)
}
