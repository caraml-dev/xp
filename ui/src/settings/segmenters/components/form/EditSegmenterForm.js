import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@caraml-dev/ui-lib";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import { ConstraintsStep } from "settings/segmenters/components/form/steps/ConstraintsStep";
import { OptionsStep } from "settings/segmenters/components/form/steps/OptionsStep";
import { SegmenterGeneralStep } from "settings/segmenters/components/form/steps/SegmenterGeneralStep";
import schema from "settings/segmenters/components/form/validation/schema";

export const EditSegmenterForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: segmenter } = useContext(FormContext);

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/segmenters/${segmenter.name}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () =>
    submitForm({ body: JSON.stringify(segmenter) }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-update-segmenter",
        title: "Segmenter updated!",
        color: "success",
        iconType: "check",
      });
      onSuccess(submissionResponse.data.name);
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "General",
      iconType: "apmTrace",
      children: <SegmenterGeneralStep isEdit={true} />,
      validationSchema: validationSchema[0],
    },
    {
      title: "Options",
      iconType: "indexSettings",
      children: <OptionsStep />,
      validationSchema: validationSchema[1],
    },
    {
      title: "Constraints",
      iconType: "fold",
      children: <ConstraintsStep />,
      validationSchema: validationSchema[2],
    },
  ];

  return (
    <AccordionForm
      name="Edit Segmenter"
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
