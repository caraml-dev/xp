import React from "react";

import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

import { Panel } from "components/panel/Panel";

export const SingleFieldConfigPanel = ({
  toolTipLabel,
  toolTipContent,
  textValue,
  textPlaceHolder,
  onChange,
  errors,
}) => {
  return (
    <Panel>
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label={toolTipLabel}
                  content={toolTipContent}
                />
              }
              isInvalid={errors?.length > 0}
              error={errors}
              display="row">
              <EuiFieldText
                fullWidth
                placeholder={textPlaceHolder}
                value={textValue}
                onChange={(e) => onChange(e.target.value)}
                isInvalid={errors?.length > 0}
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="m" />
      </EuiForm>
    </Panel>
  );
};
