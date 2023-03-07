package appcontext

import (
	"log"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/caraml-dev/xp/management-service/config"
	mw "github.com/caraml-dev/xp/management-service/middleware"
	"github.com/caraml-dev/xp/management-service/services"
	"github.com/caraml-dev/xp/management-service/services/messagequeue"
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
	log.Println("Initializing message queue publisher...")
	messageQueueService, err := messagequeue.NewMessageQueueService(*cfg.MessageQueueConfig)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing segmenter service...")
	segmenterSvc, err := services.NewSegmenterService(&allServices, cfg.SegmenterConfig, db)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing validation service...")
	validationService, err := services.NewValidationService(cfg.ValidationConfig)
	if err != nil {
		return nil, err
	}

	log.Println("Initializing experiment history service...")
	experimentHistorySvc := services.NewExperimentHistoryService(db)
	log.Println("Initializing experiment service...")
	experimentSvc := services.NewExperimentService(&allServices, db)
	log.Println("Initializing project settings service...")
	projectSettingsSvc := services.NewProjectSettingsService(&allServices, db)

	log.Println("Initializing segment history service...")
	segmentHistorySvc := services.NewSegmentHistoryService(db)
	log.Println("Initializing segment service...")
	segmentSvc := services.NewSegmentService(&allServices, db)

	log.Println("Initializing treatment history service...")
	treatmentHistorySvc := services.NewTreatmentHistoryService(db)
	log.Println("Initializing treatment service...")
	treatmentSvc := services.NewTreatmentService(&allServices, db)

	mlpSvc, err := services.NewMLPService(cfg.MLPConfig.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed initializing MLP Service")
	}

	log.Println("Initializing configuration service...")
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
		messageQueueService,
		configurationSvc,
	)

	appContext := &AppContext{
		Authorizer:       authorizer,
		OpenAPIValidator: oapiValidator,
		Services:         allServices,
	}

	return appContext, nil
}
