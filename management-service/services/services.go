package services

type Services struct {
	ExperimentService        ExperimentService
	ExperimentHistoryService ExperimentHistoryService
	SegmenterService         SegmenterService
	MLPService               MLPService
	ProjectSettingsService   ProjectSettingsService
	SegmentService           SegmentService
	SegmentHistoryService    SegmentHistoryService
	TreatmentService         TreatmentService
	TreatmentHistoryService  TreatmentHistoryService
	ValidationService        ValidationService
	MessageQueueService      MessageQueueService
	ConfigurationService     ConfigurationService
}

func NewServices(
	expSvc ExperimentService,
	expHistorySvc ExperimentHistoryService,
	segmenterSvc SegmenterService,
	mlpSvc MLPService,
	projectSettingsSvc ProjectSettingsService,
	segmentSvc SegmentService,
	segmentHistorySvc SegmentHistoryService,
	treatmentSvc TreatmentService,
	treatmentHistorySvc TreatmentHistoryService,
	validationSvc ValidationService,
	messageQueueSvc MessageQueueService,
	configurationService ConfigurationService,
) Services {
	return Services{
		ExperimentService:        expSvc,
		ExperimentHistoryService: expHistorySvc,
		MLPService:               mlpSvc,
		ProjectSettingsService:   projectSettingsSvc,
		MessageQueueService:      messageQueueSvc,
		SegmenterService:         segmenterSvc,
		SegmentService:           segmentSvc,
		SegmentHistoryService:    segmentHistorySvc,
		TreatmentService:         treatmentSvc,
		TreatmentHistoryService:  treatmentHistorySvc,
		ValidationService:        validationSvc,
		ConfigurationService:     configurationService,
	}
}
