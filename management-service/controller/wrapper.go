package controller

type Wrapper struct {
	*ProjectSettingsController
	*ExperimentController
	*ExperimentHistoryController
	*SegmentController
	*SegmentHistoryController
	*SegmenterController
	*TreatmentController
	*TreatmentHistoryController
	*ValidationController
	*ConfigurationController
}

func NewWrapper(
	settings *ProjectSettingsController,
	experiment *ExperimentController,
	experimentHistory *ExperimentHistoryController,
	segment *SegmentController,
	segmentHistory *SegmentHistoryController,
	segmenter *SegmenterController,
	treatment *TreatmentController,
	treatmentHistory *TreatmentHistoryController,
	validation *ValidationController,
	configuration *ConfigurationController,
) Wrapper {
	return Wrapper{
		ProjectSettingsController:   settings,
		ExperimentController:        experiment,
		ExperimentHistoryController: experimentHistory,
		SegmentController:           segment,
		SegmentHistoryController:    segmentHistory,
		SegmenterController:         segmenter,
		TreatmentController:         treatment,
		TreatmentHistoryController:  treatmentHistory,
		ValidationController:        validation,
		ConfigurationController:     configuration,
	}
}
