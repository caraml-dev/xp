import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@gojek/mlp-ui";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenters/context";

import { ConfigurationStep } from "./steps/ConfigurationStep";
import schema from "./validation/schema";

export const EditSegmentForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: segment } = useContext(FormContext);
  const { segmenterConfig } = useContext(SegmenterContext);
  const requiredSegmenterNames = useMemo(
    () =>
      segmenterConfig
        .filter((segment) => segment.required === true)
        .map((e) => e.name),
    [segmenterConfig]
  );

  const [submissionResponse, submitForm] = useXpApi(
    `/projects/${projectId}/segments/${segment.id}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: segment.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-edit-segment",
        title: "Segment updated!",
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
      children: <ConfigurationStep projectId={projectId} isEdit={true} />,
      validationSchema: validationSchema[0],
      validationContext: { requiredSegmenterNames },
    },
  ];

  return (
    <AccordionForm
      name="Update Segment"
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
