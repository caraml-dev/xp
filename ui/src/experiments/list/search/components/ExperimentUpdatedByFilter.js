import React from "react";

import { EuiFieldText, EuiFormRow } from "@elastic/eui";

const ExperimentUpdatedByFilter = ({ value, onChange }) => {
  return (
    <EuiFormRow fullWidth label="Updated By">
      <EuiFieldText
        compressed
        fullWidth
        placeholder="Input name"
        value={value === undefined ? "" : value}
        onChange={(e) => onChange(e.target.value)}
        name="updated-by"
      />
    </EuiFormRow>
  );
};

export default ExperimentUpdatedByFilter;
