import React from "react";

import { EuiFieldText, EuiFlexItem, EuiForm, EuiFormRow } from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

import { Panel } from "components/panel/Panel";

export const ConfigPanel = ({ name, onChange, isEdit, errors = {} }) => (
  <Panel title="General">
    <EuiForm>
      <EuiFlexItem>
        <EuiFormRow
          fullWidth
          label={<FormLabelWithToolTip label="Name *" />}
          isInvalid={!!errors.name}
          error={errors.name}
          display="row">
          <EuiFieldText
            fullWidth
            placeholder="Enter Segment Name"
            value={name}
            onChange={(e) => onChange("name")(e.target.value)}
            isInvalid={!!errors.name}
            name="segment-name"
            disabled={isEdit}
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiForm>
  </Panel>
);
