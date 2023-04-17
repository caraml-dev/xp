import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@caraml-dev/ui-lib";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenter/context";

import { RandomizationStep } from "./steps/RandomizationStep";
import { SegmentersStep } from "./steps/SegmentersStep";
import schema from "./validation/schema";

export const EditSettingsForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: settings } = useContext(FormContext);
  const { dependencyMap } = useContext(SegmenterContext);

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/settings`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: settings.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-update-settings",
        title: "Settings updated!",
        color: "success",
        iconType: "check",
      });
      onSuccess();
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "Randomization",
      iconType: "bolt",
      children: <RandomizationStep projectId={projectId} />,
      validationSchema: validationSchema[0],
    },
    {
      title: "Segmenters",
      iconType: "package",
      children: <SegmentersStep projectId={projectId} />,
      validationSchema: validationSchema[1],
      validationContext: { dependencyMap },
    },
  ];

  return (
    <AccordionForm
      name="Update Settings"
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
