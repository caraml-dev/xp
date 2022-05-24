import React from "react";

import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

import SuperSelectWithDescription from "components/form/select/SuperSelectWithDescription";
import { Panel } from "components/panel/Panel";

export const GeneralSettingsConfigPanel = ({
  name,
  type,
  tier,
  description,
  typeOptions,
  tierOptions,
  isEdit,
  onChange,
  errors = {},
}) => (
  <Panel title="General">
    <EuiForm>
      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Name *"
            content="Specify the experiment name. The name cannot be changed after creation."
          />
        }
        isInvalid={!!errors.name}
        error={errors.name}
        display="row">
        <EuiFieldText
          fullWidth
          placeholder="experiment-name"
          value={name}
          onChange={(e) => onChange("name")(e.target.value)}
          isInvalid={!!errors.name}
          disabled={isEdit}
          name="experiment-name"
        />
      </EuiFormRow>

      <EuiSpacer size="m" />

      <EuiFlexGroup direction="row">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Experiment Type *"
                content="Select the type of the experiment. The type cannot be changed after creation."
              />
            }
            isInvalid={!!errors.type}
            error={errors.type}
            display="row">
            <SuperSelectWithDescription
              fullWidth
              value={type}
              onChange={onChange("type")}
              options={typeOptions}
              hasDividers
              isInvalid={!!errors.type}
              disabled={isEdit}
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Experiment Tier *"
                content="Select the tier of the experiment. Override tier has higher priority."
              />
            }
            isInvalid={!!errors.tier}
            error={errors.tier}
            display="row">
            <SuperSelectWithDescription
              fullWidth
              value={tier}
              onChange={onChange("tier")}
              options={tierOptions}
              hasDividers
              isInvalid={!!errors.tier}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer size="m" />

      <EuiFormRow
        fullWidth
        label={
          <FormLabelWithToolTip
            label="Description"
            content="Detailed description of the experiment."
          />
        }
        isInvalid={!!errors.description}
        error={errors.description}
        display="row">
        <EuiTextArea
          fullWidth
          compressed
          value={description}
          onChange={(e) => onChange("description")(e.target.value)}
          isInvalid={!!errors.description}
          resize="vertical"
        />
      </EuiFormRow>
    </EuiForm>
  </Panel>
);
