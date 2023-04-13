import React, { useEffect, useState } from "react";

import {
  EuiComboBox,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

import { useXpApi } from "hooks/useXpApi";

export const TreatmentConfigPanel = ({
  projectId,
  treatmentConfig,
  treatmentTemplate,
  treatmentSelectionOptions,
  onChange,
  errors = {},
}) => {
  const [treatmentId, setTreatmentId] = useState();
  const [hasNewResponse, setHasNewResponse] = useState(false);

  const [
    { data: treatmentDetails, isLoaded: isAPILoaded },
    fetchTreatmentDetails,
  ] = useXpApi(
    `/projects/${projectId}/treatments/${treatmentId}`,
    {},
    {},
    false
  );

  const onCustomOrTemplateSelection = (selected) => {
    onChange("configuration")("");

    // If id is not present, it will be set to undefined, but will not trigger useEffect
    setTreatmentId(selected[0]?.id);
    onChange("treatment_template")(selected.length > 0 ? selected[0] : "");
  };

  // Fetch Treatment details every time there is a selection of Treatment with id
  useEffect(() => {
    if (!!treatmentId) {
      fetchTreatmentDetails();
      setHasNewResponse(true);
    }
  }, [treatmentId, fetchTreatmentDetails]);

  // Populate Treatment object so that it will reflect on UI
  useEffect(() => {
    if (hasNewResponse && isAPILoaded) {
      onChange("configuration")(
        JSON.stringify(treatmentDetails.data.configuration)
      );
      setHasNewResponse(false);
    }
  }, [onChange, hasNewResponse, isAPILoaded, treatmentDetails]);

  return (
    <EuiForm>
      <EuiFormRow fullWidth label="Template">
        <EuiComboBox
          placeholder="Copy from Pre-configured Treatment"
          isDisabled={
            !treatmentSelectionOptions || treatmentSelectionOptions.length === 0
          }
          fullWidth={true}
          singleSelection={{ asPlainText: true }}
          options={treatmentSelectionOptions}
          onChange={onCustomOrTemplateSelection}
          selectedOptions={!!treatmentTemplate ? [treatmentTemplate] : []}
        />
      </EuiFormRow>
      <EuiSpacer />
      <EuiFlexGroup direction="column">
        <EuiFlexItem grow={1}>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Configuration"
                content="Specify the configuration for the Treatment."
              />
            }
            isInvalid={!!errors.configuration}
            error={errors.configuration}
            display="row">
            <EuiTextArea
              fullWidth
              placeholder="Enter valid JSON"
              value={treatmentConfig}
              onChange={(e) => onChange("configuration")(e.target.value)}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiForm>
  );
};
