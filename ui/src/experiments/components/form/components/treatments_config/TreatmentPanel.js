import React, { Fragment } from "react";

import {
  EuiFieldNumber,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiSpacer,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

export const TreatmentPanel = ({ errors, treatment, onUpdateTreatment }) => {
  return (
    <Fragment>
      <EuiFlexGroup direction="row">
        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Treatment Name"
                content="Specify the name of the Treatment."
              />
            }
            isInvalid={!!errors.name}
            error={errors.name}
            display="row">
            <EuiFieldText
              isInvalid={errors ? !!errors.name : false}
              value={treatment.name}
              onChange={(e) =>
                onUpdateTreatment(treatment.uuid, "name", e.target.value)
              }
            />
          </EuiFormRow>
        </EuiFlexItem>
        <EuiFlexItem>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Traffic Percentage"
                content="Specify the traffic percentage for the Treatment."
              />
            }
            isInvalid={!!errors.traffic}
            error={errors.traffic}
            display="row">
            <EuiFieldNumber
              value={treatment.traffic}
              onChange={(e) =>
                onUpdateTreatment(
                  treatment.uuid,
                  "traffic",
                  Number(e.target.value)
                )
              }
              isInvalid={false}
              min={0}
              max={100}
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer size="m" />
      <EuiFlexGroup>
        <EuiFlexItem>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Configuration"
                content="Specify the configuration for the Treatment."
              />
            }
            isInvalid={!!errors.configuration}
            error={errors.configuration}
            display="row">
            <EuiTextArea
              fullWidth
              placeholder="Enter valid JSON"
              value={treatment.configuration}
              onChange={(e) =>
                onUpdateTreatment(
                  treatment.uuid,
                  "configuration",
                  e.target.value
                )
              }
            />
          </EuiFormRow>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Fragment>
  );
};
