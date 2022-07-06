import React from "react";

import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";

import { Panel } from "components/panel/Panel";
import { TreatmentValidationRulesCard } from "settings/components/form/components/treatment_validation_section/TreatmentValidationRulesCard";

export const TreatmentValidationRulesConfigPanel = ({
  settings,
  onChange,
  errors = [],
}) => {
  var rules = settings?.treatment_schema?.rules || [];

  const onAddRule = () => {
    rules.push({
      name: "",
      predicate: "",
    });
    onChange(`treatment_schema.rules`)(rules);
  };

  return (
    <Panel title="Treatment Validation Rules">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="s">
        {rules.map((rule, idx) => (
          <EuiFlexItem key={idx}>
            <TreatmentValidationRulesCard
              rule={rule}
              onChangeHandler={onChange(`treatment_schema.rules[${idx}]`)}
              onDeleteRule={() => {
                rules.splice(idx, 1);
                onChange(`treatment_schema.rules`)(rules);
              }}
              errors={errors?.treatment_schema?.rules[idx]}
            />
          </EuiFlexItem>
        ))}
      </EuiFlexGroup>
      <EuiSpacer size="m" />
      <EuiFlexItem>
        <EuiButton onClick={onAddRule}>+ Add Rule</EuiButton>
      </EuiFlexItem>
    </Panel>
  );
};
