import React, { useContext, useEffect, useMemo } from "react";

import { AccordionForm, FormContext, addToast } from "@gojek/mlp-ui";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenter/context";

import { ConfigurationStep } from "./steps/ConfigurationStep";
import schema from "./validation/schema";

export const CreateSegmentForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: segment } = useContext(FormContext);
  const { segmenterConfig, getSegmenterOptions } = useContext(SegmenterContext);

  // retrieve name-type mappings for active segmenter specified for this project
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
    `/projects/${projectId}/segments`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );
  const onSubmit = () => submitForm({ body: segment.stringify() }).promise;

  useEffect(() => {
    if (submissionResponse.isLoaded && !submissionResponse.error) {
      addToast({
        id: "submit-success-create-segment",
        title: "Segment created!",
        color: "success",
        iconType: "check",
      });
      onSuccess(submissionResponse.data.data.id);
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "Segments",
      iconType: "package",
      children: <ConfigurationStep projectId={projectId} isEdit={false} />,
      validationSchema: validationSchema[0],
      validationContext: { requiredSegmenterNames, segmenterTypes },
    },
  ];

  return (
    <AccordionForm
      name="Create Segment"
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
