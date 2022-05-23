import React from "react";

import { EuiCheckbox, EuiFormRow } from "@elastic/eui";

const ExperimentSegmenterMatchOptionsFilter = ({ value, onChange }) => (
  <EuiFormRow fullWidth label="Segmenter Match Options">
    <EuiCheckbox
      id={"include_weak_match"}
      label={"Include weak matches"}
      checked={!!value}
      onChange={() => onChange(!value)}
      compressed
    />
  </EuiFormRow>
);

export default ExperimentSegmenterMatchOptionsFilter;
