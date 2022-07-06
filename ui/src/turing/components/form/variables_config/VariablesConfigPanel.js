import React, { useContext, useEffect } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { FormLabelWithToolTip, get, useOnChangeHandler } from "@gojek/mlp-ui";

import { Panel } from "components/panel/Panel";
import SettingsContext from "providers/settings/context";
import { VariableConfigRow } from "turing/components/form/variables_config/VariableConfigRow";

export const VariablesConfigPanel = ({
  variables = [],
  onChangeHandler,
  errors = {},
}) => {
  const { variables: allVars, isLoaded } = useContext(SettingsContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  // Update config using the onchange handler if anything changed in the settings
  useEffect(() => {
    if (isLoaded("variables")) {
      const existingVars = variables.map((e) => e.name);
      const missingVars = allVars.filter((e) => !existingVars.includes(e));
      const extraVars = existingVars.filter((e) => !allVars.includes(e));
      // Add vars that are missing and remove the ones not required
      const updatedVars = [
        ...variables,
        ...missingVars.map((e) => ({
          name: e,
          field: "",
          field_source: "header",
        })),
      ].filter((e) => !extraVars.includes(e.name));
      if (missingVars.length > 0 || extraVars.length > 0) {
        // Something changed, call onchange handler
        onChangeHandler(updatedVars);
      }
    }
  }, [allVars, isLoaded, variables, onChangeHandler]);

  return (
    <Panel
      title={
        <FormLabelWithToolTip
          label="Variables"
          size="m"
          content="Specify how the experiment variables may be parsed from the request."
        />
      }>
      <></>
      <EuiSpacer size="xs" />
      <EuiFlexItem>
        <EuiFlexGroup direction="column" gutterSize="s">
          {variables.map((variable, idx) => (
            <EuiFlexItem key={`experiment-variable-${variable.name}`}>
              <VariableConfigRow
                name={variable.name}
                field={variable.field}
                fieldSrc={variable.field_source}
                onChangeHandler={onChange(`${idx}`)}
                error={get(errors, `${idx}`)}
              />
            </EuiFlexItem>
          ))}
        </EuiFlexGroup>
      </EuiFlexItem>
    </Panel>
  );
};
