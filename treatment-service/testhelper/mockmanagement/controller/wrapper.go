package controller

type Wrapper struct {
	ProjectSettings
	Experiment
	ExperimentHistory
	Segmenter
}

func NewWrapper(
	settings ProjectSettings,
	experiment Experiment,
	experimentHistory ExperimentHistory,
	segmenter Segmenter,
) Wrapper {
	return Wrapper{
		ProjectSettings:   settings,
		Experiment:        experiment,
		ExperimentHistory: experimentHistory,
		Segmenter:         segmenter,
	}
}
