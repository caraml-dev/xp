import React from "react";

import {
  EuiCheckbox,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiSuperSelect,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

import { Panel } from "components/panel/Panel";
import { typeOptions } from "settings/segmenters/components/typeOptions";

export const SegmenterGeneralPanel = ({
  name,
  type,
  description,
  required,
  multiValued,
  onChange,
  isEdit,
  errors = {},
}) => (
  <Panel title="General">
    <EuiForm>
      <EuiFlexGroup direction="row">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Name *"
                content="Specify the segmenter name. This name has to be unique across all segmenters."
              />
            }
            isInvalid={!!errors.name}
            error={errors.name}
            display="row">
            <EuiFieldText
              fullWidth
              placeholder="Enter Segmenter Name"
              value={name}
              onChange={(e) => onChange("name")(e.target.value)}
              isInvalid={!!errors.name}
              disabled={isEdit}
              name="segmenter-name"
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Type *"
                content="Select the type of the segmenter. The type cannot be changed after creation."
              />
            }
            isInvalid={!!errors.type}
            error={errors.type}
            display="row">
            <EuiSuperSelect
              fullWidth
              options={typeOptions}
              valueOfSelected={type}
              onChange={onChange("type")}
              isInvalid={!!errors.type}
              disabled={isEdit}
              itemLayoutAlign="top"
              hasDividers
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
            content="Detailed description of the segmenter."
          />
        }
        isInvalid={!!errors.description}
        error={errors.description}
        display="row">
        <EuiFieldText
          fullWidth
          placeholder="Enter Segmenter Description"
          value={description}
          onChange={(e) => onChange("description")(e.target.value)}
          isInvalid={!!errors.description}
          name="segmenter-description"
        />
      </EuiFormRow>

      <EuiSpacer size="m" />

      <EuiFlexGroup direction="row">
        <EuiFlexItem>
          <EuiCheckbox
            id="requiredCheckbox"
            label={
              <FormLabelWithToolTip
                label="Required"
                content="Specify if this segmenter is always required to be active."
              />
            }
            checked={required}
            onChange={(e) => onChange("required")(e.target.checked)}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiCheckbox
            id="multiValuedCheckbox"
            label={
              <FormLabelWithToolTip
                label="Multi-Valued"
                content="If selected, multiple values of the segmenter can be chosen at once, in segment configurations."
              />
            }
            checked={multiValued}
            onChange={(e) => onChange("multi_valued")(e.target.checked)}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </EuiForm>
  </Panel>
);
