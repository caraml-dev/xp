import React from "react";

import { EuiCheckboxGroup } from "@elastic/eui";

const SingleSelectCheckboxGroup = ({
  options,
  currentValue,
  compressed,
  onChange,
}) => {
  // Assign ids.
  const optionsIdMap = options.reduce((acc, e) => {
    const id = e.value; // Use the value as id, as they are unique
    acc[id] = { ...e, label: e.label || e.value, id };
    return acc;
  }, {});

  const idToSelectedMap = Object.values(optionsIdMap).reduce((acc, e) => {
    acc[e.id] = e.value === currentValue;
    return acc;
  }, {});

  // If the changed value from the checkbox group is already selected, clear the selection.
  // Else, use that as the new selection.
  const setOrClearValue = (newValue) =>
    newValue === currentValue ? onChange() : onChange(newValue);

  return (
    <EuiCheckboxGroup
      options={Object.values(optionsIdMap)}
      idToSelectedMap={idToSelectedMap}
      onChange={(id) => setOrClearValue(optionsIdMap[id].value)}
      compressed={compressed}
    />
  );
};

export default SingleSelectCheckboxGroup;
