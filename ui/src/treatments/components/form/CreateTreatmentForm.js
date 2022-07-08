import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@gojek/mlp-ui";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";

import { ConfigurationStep } from "./steps/ConfigurationStep";
import schema from "./validation/schema";

export const CreateTreatmentForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: treatment } = useContext(FormContext);

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/treatments`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: treatment.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-create-treatment",
        title: "Treatment created!",
        color: "success",
        iconType: "check",
      });
      onSuccess(submissionResponse.data.data.id);
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "General",
      iconType: "apmTrace",
      children: <ConfigurationStep projectId={projectId} />,
      validationSchema: validationSchema[0],
    },
  ];

  return (
    <AccordionForm
      name="Create Treatment"
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
