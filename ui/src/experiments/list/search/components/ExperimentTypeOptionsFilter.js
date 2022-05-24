import React from "react";

import { EuiFormRow } from "@elastic/eui";

import SingleSelectCheckboxGroup from "components/form/checkbox/SingleSelectCheckboxGroup";

const ExperimentTypeOptionsFilter = ({ value, onChange, options, label }) => (
  <EuiFormRow fullWidth label={label}>
    <SingleSelectCheckboxGroup
      options={options}
      currentValue={value}
      onChange={onChange}
      compressed
    />
  </EuiFormRow>
);

export default ExperimentTypeOptionsFilter;
