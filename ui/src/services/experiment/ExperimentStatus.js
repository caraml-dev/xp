import { experimentStatusesFriendly } from "experiments/components/typeOptions";

export const getExperimentStatus = (experiment) => experimentStatusesFriendly.find(e => e.value === experiment.status_friendly);
