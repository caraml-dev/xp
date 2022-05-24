import React, { useContext, useEffect, useMemo } from "react";

import { EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import { AccordionForm, FormContext, addToast } from "@gojek/mlp-ui";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenters/context";
import SettingsContext from "providers/settings/context";

import { GeneralStep } from "./steps/GeneralStep";
import { SegmentStep } from "./steps/SegmentStep";
import { TreatmentsStep } from "./steps/TreatmentsStep";
import schema from "./validation/schema";

export const EditExperimentForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { settings, isLoaded } = useContext(SettingsContext);
  const { data: experiment } = useContext(FormContext);
  const { segmenterConfig } = useContext(SegmenterContext);
  const requiredSegmenterNames = useMemo(
    () =>
      segmenterConfig
        .filter((segment) => segment.required === true)
        .map((e) => e.name),
    [segmenterConfig]
  );

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/experiments/${experiment.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  const onSubmit = () => {
    for (const key of Object.keys(experiment.segment)) {
      if (!settings.segmenters.names.includes(key)) {
        delete experiment.segment[key];
      }
    }
    return submitForm({ body: experiment.stringify() }).promise;
  };

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-experiment",
        title: "Experiment updated!",
        color: "success",
        iconType: "check",
      });
      onSuccess();
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "General",
      iconType: "apmTrace",
      children: <GeneralStep projectId={projectId} />,
      validationSchema: validationSchema[0],
    },
    {
      title: "Segment",
      iconType: "package",
      children: <SegmentStep projectId={projectId} isEdit={false} />,
      validationSchema: validationSchema[1],
      validationContext: { requiredSegmenterNames },
    },
    {
      title: "Treatments",
      iconType: "beaker",
      children: <TreatmentsStep projectId={projectId} />,
      validationSchema: validationSchema[2],
    },
  ];

  if (isLoaded) {
    if (settings && settings.segmenters.names.length === 0) {
      sections.splice(1, 1);
    }
  }

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : (
    <AccordionForm
      name="Edit Experiment"
      sections={sections}
      onCancel={onCancel}
      onSubmit={onSubmit}
      submitLabel="Save"
      renderTitle={(title, iconType) => (
        <ConfigSectionTitle title={title} iconType={iconType} />
      )}
    />
  );
};
