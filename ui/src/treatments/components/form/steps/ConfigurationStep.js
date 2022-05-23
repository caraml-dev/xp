import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@gojek/mlp-ui";

import { ConfigPanel } from "treatments/components/form/components/config_section/ConfigPanel";
import { TreatmentSelectionPanel } from "treatments/components/form/components/config_section/TreatmentSelectionPanel";

export const ConfigurationStep = ({ projectId }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <ConfigPanel
          name={data.name}
          onChange={onChange}
          errors={errors}
          isEdit={!!data.id}
        />
        <EuiSpacer />
        <TreatmentSelectionPanel
          projectId={projectId}
          treatmentConfig={data.configuration}
          treatmentTemplate={data.treatment_template}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
