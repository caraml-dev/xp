import React, { useEffect, useState } from "react";

import { EuiComboBox, EuiFormRow } from "@elastic/eui";

const ExperimentSegmenterFilter = ({
  name,
  filteredOptions,
  onChange,
  isMultiValued,
  value,
}) => {
  const [selectedOptions, setSelected] = useState([]);

  useEffect(() => {
    if (value) {
      // Reset constrained set of dependent filters for display
      const trimmedOptions = filteredOptions.filter((opt) => {
        return value.includes(opt.value);
      });
      setSelected(trimmedOptions);
    } else {
      setSelected([]);
    }
  }, [value, filteredOptions]);

  const onChangeOpt = (selectedOptions) => {
    setSelected(selectedOptions);

    /* Trigger hot reload on
    - filter options panel with selected segmenter filters
    - filtered table with selected segmenter filters
    */
    onChange(selectedOptions.map((opt) => opt.value));
  };

  return (
    <EuiFormRow fullWidth label={name}>
      <EuiComboBox
        placeholder="Select options"
        options={filteredOptions}
        selectedOptions={selectedOptions}
        onChange={onChangeOpt}
        isClearable={true}
        data-test-subj={`${name}-segmenter-combo-box`}
        singleSelection={!isMultiValued}
      />
    </EuiFormRow>
  );
};

export default ExperimentSegmenterFilter;
