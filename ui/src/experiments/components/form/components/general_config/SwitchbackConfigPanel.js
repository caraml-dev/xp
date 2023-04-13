import React from "react";

import {
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

import { Panel } from "components/panel/Panel";

export const SwitchbackConfigPanel = ({ interval, onChange, errors = {} }) => (
  <Panel title="Switchback Configuration">
    <EuiForm>
      <EuiFlexGroup direction="row">
        <EuiFlexItem grow={true}>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Switchback Interval (Minutes) *"
                content="Specify the Switchback Interval in minutes. The minimum allowed value is 5."
              />
            }
            isInvalid={!!errors.interval}
            error={errors.interval}
            display="row">
            <EuiFieldNumber
              fullWidth
              placeholder="Specify the interval"
              value={interval}
              onChange={(e) => onChange("interval")(Number(e.target.value))}
              isInvalid={!!errors.interval}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiForm>
  </Panel>
);
