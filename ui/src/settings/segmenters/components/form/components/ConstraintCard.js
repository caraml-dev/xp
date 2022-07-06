import React from "react";

import {
  EuiButtonIcon,
  EuiCard,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip, get, useOnChangeHandler } from "@gojek/mlp-ui";

export const ConstraintCard = ({
  constraint,
  onChangeHandler,
  onDelete,
  errors,
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const getPreRequisiteErrors = function (errors) {
    var error = get(errors, "pre_requisites");

    // return null error or any top level error for pre_requisites
    if (!error || Array.isArray(error)) return error;

    for (const [, preRequisiteError] of Object.entries(error)) {
      if (
        preRequisiteError.segmenter_name &&
        Array.isArray(preRequisiteError.segmenter_name)
      )
        return preRequisiteError.segmenter_name;
      if (
        preRequisiteError.segmenter_values &&
        Array.isArray(preRequisiteError.segmenter_values)
      )
        return preRequisiteError.segmenter_values;
    }
    return null;
  };

  return (
    <EuiCard title="" textAlign="left">
      <EuiFlexGroup justifyContent="flexEnd" gutterSize="none" direction="row">
        <EuiFlexItem grow={false}>
          <EuiButtonIcon
            iconType="cross"
            onClick={onDelete}
            aria-label="delete-constraint"
          />
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiFlexGroup direction="row">
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Pre-Requisite Segmenter Values *"
                content="Specify the pre-requisite segmenter values."
              />
            }
            isInvalid={!!getPreRequisiteErrors(errors)}
            error={getPreRequisiteErrors(errors)}
            display="row">
            <EuiTextArea
              fullWidth
              placeholder={`Enter pre-requisite segmenter values. Eg: \n${JSON.stringify(
                [
                  {
                    segmenter_name: "country_code",
                    segmenter_values: ["ID"],
                  },
                ],
                null,
                4
              )}`}
              value={constraint.pre_requisites}
              onChange={(e) => onChange("pre_requisites")(e.target.value)}
              isInvalid={!!getPreRequisiteErrors(errors)}
              name="segmenter-constraint-pre-requisites"
            />
          </EuiFormRow>
        </EuiFlexItem>

        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Allowed Values *"
                content="Specify the values the segmenter should allow. Values specified here should form
                  a subset of the values specified in the options field."
              />
            }
            isInvalid={!!get(errors, "allowed_values")}
            error={get(errors, "allowed_values")}
            display="row">
            <EuiTextArea
              fullWidth
              placeholder={`Enter allowed values. Eg: \n${JSON.stringify(
                [0, 1],
                null,
                4
              )}`}
              value={constraint.allowed_values}
              onChange={(e) => onChange("allowed_values")(e.target.value)}
              isInvalid={!!get(errors, "allowed_values")}
              name="segmenter-constraint-allowed-values"
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiSpacer size="m" />

      <EuiFlexItem>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Value Overrides"
              content="Specify the name-value pairs to overwrite the names for the values under 'Options'. This field
                should specify the name-value mappings for each and every value in the allowed values field. Values
                specified here should form a subset of the values specified in the options field."
            />
          }
          isInvalid={!!get(errors, "options")}
          error={get(errors, "options")}
          display="row">
          <EuiTextArea
            fullWidth
            placeholder={`Enter valid JSON. Eg: \n${JSON.stringify(
              {
                ID_BIKE: 0,
                ID_CAR: 1,
              },
              null,
              4
            )}`}
            value={constraint.options}
            onChange={(e) => onChange("options")(e.target.value)}
            isInvalid={!!get(errors, "options")}
            name="segmenter-constraint-options"
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiCard>
  );
};
