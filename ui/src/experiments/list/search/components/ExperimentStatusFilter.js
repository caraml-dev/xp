import React from "react";

import { EuiComboBox, EuiFormRow, EuiHealth } from "@elastic/eui";

const ExperimentStatusFilter = ({
  value,
  options,
  onChange,
}) => {
  const renderOption = o => (<EuiHealth color={o.color}>{o.label}</EuiHealth>);
  const selectedOptions = options.filter(o => !!value && value.includes(o.value));
  // Filter only fields recognized by the combobox.
  const optionsFiltered = options.map(o => ({ value: o.value, label: o.label, color: o.color }));

  return (
    <EuiFormRow fullWidth label={"Experiment Status"}>
      <EuiComboBox
        placeholder="Select option"
        options={optionsFiltered}
        renderOption={renderOption}
        selectedOptions={selectedOptions}
        onChange={e => onChange(e.map(item => item.value))}
        isClearable={true}
      />
    </EuiFormRow>
  )
};

export default ExperimentStatusFilter;
