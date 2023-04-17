import React from "react";

import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormLabel,
  EuiFormRow,
} from "@elastic/eui";
import { useOnChangeHandler } from "@caraml-dev/ui-lib";

import { FieldSourceFormLabel } from "./components/FieldSourceFormLabel";

import "./VariableConfigRow.scss";

export const VariableConfigRow = ({
  name,
  field,
  fieldSrc,
  protocol,
  onChangeHandler,
  error = {},
}) => {
  // Define onChange handlers
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const onChangeFieldSource = (newValue) => {
    onChange("field_source")(newValue);
    // Clear the field value if field source is none
    if (newValue === "none") {
      onChange("field")("");
    }
  };

  return (
    <EuiFlexGroup
      direction="row"
      gutterSize="m"
      alignItems="center"
      className="euiFlexGroup--experimentVariableRow"
    >
      <EuiFlexItem grow={true} className="eui-textTruncate">
        <EuiFormRow
          fullWidth
          isInvalid={!!error.field}
          error={error.field || ""}
        >
          <EuiFieldText
            fullWidth
            compressed
            placeholder={"Enter Field Name..."}
            value={field || ""}
            onChange={(e) => onChange("field")(e.target.value)}
            isInvalid={!!error.field}
            disabled={fieldSrc === "none"}
            prepend={
              <FieldSourceFormLabel
                readOnly={false}
                protocol={protocol}
                value={fieldSrc}
                onChange={(e) => onChangeFieldSource(e)}
              />
            }
            append={<EuiFormLabel>{name}</EuiFormLabel>}
          />
        </EuiFormRow>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
