package appcontext

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/caraml-dev/xp/management-service/config"
	mw "github.com/caraml-dev/xp/management-service/middleware"
	"github.com/caraml-dev/xp/management-service/services"
)

type AppContext struct {
	Authorizer       *mw.Authorizer
	OpenAPIValidator *mw.OpenAPIValidator
	Services         services.Services
}

func NewAppContext(db *gorm.DB, authorizer *mw.Authorizer, cfg *config.Config) (*AppContext, error) {
	// Init Services
	var allServices services.Services

	// Init Validator
	oapiValidator, err := mw.NewOpenAPIValidator(&mw.OpenAPIValidationOptions{
		IgnoreAuthentication: true,
		IgnoreServers:        true,
	})
	if err != nil {
		return nil, err
	}

	// Init Services
	pubSubConfig := config.PubSubConfig{
		Project:   cfg.PubSubConfig.Project,
		TopicName: cfg.PubSubConfig.TopicName,
	}
	pubSubPublisherService, err := services.NewPubSubPublisherService(&pubSubConfig)
	if err != nil {
		return nil, err
	}

	segmenterSvc, err := services.NewSegmenterService(&allServices, cfg.SegmenterConfig, db)
	if err != nil {
		return nil, err
	}

	validationService, err := services.NewValidationService(cfg.ValidationConfig)
	if err != nil {
		return nil, err
	}

	experimentHistorySvc := services.NewExperimentHistoryService(db)
	experimentSvc := services.NewExperimentService(&allServices, db)
	projectSettingsSvc := services.NewProjectSettingsService(&allServices, db)

	segmentHistorySvc := services.NewSegmentHistoryService(db)
	segmentSvc := services.NewSegmentService(&allServices, db)

	treatmentHistorySvc := services.NewTreatmentHistoryService(db)
	treatmentSvc := services.NewTreatmentService(&allServices, db)

	mlpSvc, err := services.NewMLPService(cfg.MLPConfig.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing MLP Service")
	}

	configurationSvc := services.NewConfigurationService(cfg)

	allServices = services.NewServices(
		experimentSvc,
		experimentHistorySvc,
		segmenterSvc,
		mlpSvc,
		projectSettingsSvc,
		segmentSvc,
		segmentHistorySvc,
		treatmentSvc,
		treatmentHistorySvc,
		validationService,
		pubSubPublisherService,
		configurationSvc,
	)

	appContext := &AppContext{
		Authorizer:       authorizer,
		OpenAPIValidator: oapiValidator,
		Services:         allServices,
	}

	return appContext, nil
}
