import React, { useContext, useEffect, useMemo } from "react";

import { EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import { FormContext, StepsWizardHorizontal, addToast } from "@gojek/mlp-ui";

import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenter/context";
import SettingsContext from "providers/settings/context";

import { GeneralStep } from "./steps/GeneralStep";
import { SegmentStep } from "./steps/SegmentStep";
import { TreatmentsStep } from "./steps/TreatmentsStep";
import schema from "./validation/schema";

export const CreateExperimentForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { settings, isLoaded } = useContext(SettingsContext);
  const { data: experiment } = useContext(FormContext);
  const { segmenterConfig, getSegmenterOptions } = useContext(SegmenterContext);

  // retrieve name-type mappings for active segmenters specified for this project
  const segmenterTypes = getSegmenterOptions(segmenterConfig).reduce(function(
    map,
    obj
  ) {
    map[obj.name] = obj.type;
    return map;
  },
    {});

  const requiredSegmenterNames = useMemo(
    () =>
      segmenterConfig
        .filter((segment) => segment.required === true)
        .map((e) => e.name),
    [segmenterConfig]
  );

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: experiment.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-experiment",
        title: "New Experiment created!",
        color: "success",
        iconType: "check",
      });
      onSuccess(submissionResponse.data.data.id);
    }
  }, [submissionResponse, onSuccess]);

  const steps = [
    {
      title: "General",
      children: <GeneralStep projectId={projectId} />,
      validationSchema: validationSchema[0],
    },
    {
      title: "Segment",
      children: <SegmentStep projectId={projectId} isEdit={false} />,
      validationSchema: validationSchema[1],
      validationContext: { requiredSegmenterNames, segmenterTypes },
    },
    {
      title: "Treatments",
      children: <TreatmentsStep projectId={projectId} />,
      validationSchema: validationSchema[2],
    },
  ];

  if (isLoaded) {
    if (settings && settings.segmenters.names.length === 0) {
      steps.splice(1, 1);
    }
  }

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : (
    <StepsWizardHorizontal
      steps={steps}
      onCancel={onCancel}
      onSubmit={onSubmit}
      submitLabel="Save"
    />
  );
};
