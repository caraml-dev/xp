import React, { Fragment, useContext, useEffect, useMemo } from "react";

import { EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import { AccordionForm, FormContext, addToast } from "@caraml-dev/ui-lib";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";
import { useXpApi } from "hooks/useXpApi";
import { ExternalValidationStep } from "settings/components/form/steps/ExternalValidationStep";
import { TreatmentValidationRulesStep } from "settings/components/form/steps/TreatmentValidationRulesStep";
import schema from "settings/components/form/validation/schema";

export const EditValidationForm = ({ projectId, onCancel, onSuccess }) => {
  const validationSchema = useMemo(() => schema, []);
  const { data: settings } = useContext(FormContext);

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
        title: "Validation updated!",
        color: "success",
        iconType: "check",
      });
      onSuccess();
    }
  }, [submissionResponse, onSuccess]);

  const sections = [
    {
      title: "External Validation",
      iconType: "symlink",
      children: <ExternalValidationStep projectId={projectId} />,
      validationSchema: validationSchema[2],
    },
    {
      title: "Treatment Validation Rules",
      iconType: "inspect",
      children: <TreatmentValidationRulesStep projectId={projectId} />,
      validationSchema: validationSchema[3],
    },
  ];

  return (
    <Fragment>
      {!settings ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : (
        <AccordionForm
          name="Update Validation"
          sections={sections}
          onCancel={onCancel}
          onSubmit={onSubmit}
          submitLabel="Save"
          renderTitle={(title, iconType) => (
            <ConfigSectionTitle title={title} iconType={iconType} />
          )}
        />
      )}
    </Fragment>
  );
};
