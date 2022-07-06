import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@gojek/mlp-ui";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenters/context";
import { RandomizationStep } from "settings/components/form/steps/RandomizationStep";
import { SegmentersStep } from "settings/components/form/steps/SegmentersStep";
import schema from "settings/components/form/validation/schema";

export const CreateSettingsForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: settings } = useContext(FormContext);
  const { dependencyMap } = useContext(SegmenterContext);

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/settings/`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: settings.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-create-settings",
        title: "Settings created!",
        color: "success",
        iconType: "check",
      });
      onSuccess(submissionResponse.data.data.id);
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
      name="Create Settings"
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
