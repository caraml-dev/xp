import React from "react";

import { EuiForm, EuiFormRow, EuiTextArea } from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";

import { Panel } from "components/panel/Panel";

export const OptionsPanel = ({ options, onChange, errors = {} }) => {
  return (
    <Panel title="Options">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Options"
              content="Specify the name-value mappings for the segmenter."
            />
          }
          isInvalid={!!errors.options}
          error={errors.options}
          display="row">
          <EuiTextArea
            fullWidth
            placeholder={`Enter valid JSON. Eg: \n${JSON.stringify(
              {
                BIKE: 0,
                CAR: 1,
                TRUCK: 2,
                RUNNER: 3,
                CYCLE: 4,
              },
              null,
              4
            )}`}
            value={options}
            onChange={(e) => onChange("options")(e.target.value)}
            isInvalid={!!errors.options}
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
