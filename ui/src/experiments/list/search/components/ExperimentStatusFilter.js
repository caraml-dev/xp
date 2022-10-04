import React from "react";

import { EuiComboBox, EuiFormRow, EuiHealth } from "@elastic/eui";

const ExperimentStatusFilter = ({
  value,
  options,
  onChange,
}) => {
  const renderOption = o => (<EuiHealth color={o.color}>{o.label}</EuiHealth>);
  const selectedOptions = options.filter(o => o.value === value);
  // Filter only fields recognized by the combobox.
  const optionsFiltered = options.map(o => ({ value: o.value, label: o.label, color: o.color }));

  return (
    <EuiFormRow fullWidth label={"Experiment Status"}>
      <EuiComboBox
        placeholder="Select option"
        options={optionsFiltered}
        renderOption={renderOption}
        selectedOptions={selectedOptions}
        onChange={e => onChange(e.length > 0 ? e[0].value : undefined)}
        isClearable={true}
        data-test-subj={`experiment-status-combo-box`}
        singleSelection={true}
      />
    </EuiFormRow>
  )
};

export default ExperimentStatusFilter;
