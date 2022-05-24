import React from "react";

import {
  EuiButtonIcon,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiPanel,
  EuiSpacer,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip, useOnChangeHandler } from "@gojek/mlp-ui";

export const TreatmentValidationRulesCard = ({
  rule,
  onChangeHandler,
  onDeleteRule,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiPanel>
      <EuiForm>
        <EuiFlexGroup justifyContent="flexEnd">
          <EuiButtonIcon
            iconType="cross"
            onClick={onDeleteRule}
            aria-label="delete-treatment-rule"
          />
        </EuiFlexGroup>
        <EuiFlexGroup direction="column">
          <EuiFlexItem grow={true}>
            <EuiFormRow
              fullWidth
              label="Name"
              isInvalid={!!errors.name}
              error={errors.name}
              display="row">
              <EuiFieldText
                fullWidth
                placeholder="Enter Rule Name"
                value={rule.name}
                onChange={(e) => onChange(`name`)(e.target.value)}
                isInvalid={!!errors.name}
              />
            </EuiFormRow>
          </EuiFlexItem>
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Predicate"
                  content="A Go template expression that must return a boolean value. Besides the basic Go template functions, all Sprig library functions are supported."
                />
              }
              isInvalid={!!errors.predicate}
              error={errors.predicate}
              display="row">
              <EuiTextArea
                fullWidth
                placeholder={`Enter the predicate. Eg:\n{{- (eq .field1 "abc") -}}`}
                value={rule.predicate}
                onChange={(e) => onChange(`predicate`)(e.target.value)}
                isInvalid={!!errors.predicate}
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="m" />
      </EuiForm>
    </EuiPanel>
  );
};
