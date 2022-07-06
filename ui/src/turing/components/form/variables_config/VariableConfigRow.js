import "turing/components/form/variables_config/VariableConfigRow.scss";

import React from "react";

import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormLabel,
  EuiFormRow,
} from "@elastic/eui";
import { useOnChangeHandler } from "@gojek/mlp-ui";

import { FieldSourceFormLabel } from "turing/components/form/variables_config/components/FieldSourceFormLabel";

export const VariableConfigRow = ({
  name,
  field,
  fieldSrc,
  onChangeHandler,
  error = {},
}) => {
  // Define onChange handlers
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup
      direction="row"
      gutterSize="m"
      alignItems="center"
      className="euiFlexGroup--experimentVariableRow">
      <EuiFlexItem grow={true} className="eui-textTruncate">
        <EuiFormRow
          fullWidth
          isInvalid={!!error.field}
          error={error.field || ""}>
          <EuiFieldText
            fullWidth
            compressed
            placeholder={"Enter Field Name..."}
            value={field || ""}
            onChange={(e) => onChange("field")(e.target.value)}
            isInvalid={!!error.field}
            prepend={
              <FieldSourceFormLabel
                readOnly={false}
                value={fieldSrc}
                onChange={onChange("field_source")}
              />
            }
            append={<EuiFormLabel>{name}</EuiFormLabel>}
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
