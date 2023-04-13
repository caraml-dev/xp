import React, { useEffect, useMemo, useState } from "react";

import {
  EuiComboBox,
  EuiFieldText,
  EuiFlexItem,
  EuiFormRow,
  EuiTextArea,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";
import isEqual from "lodash/isEqual";

import { parseSegmenterValue } from "services/experiment/Segment";

const SelectableSegmenterConfigRow = ({
  isMultiValued,
  options,
  values,
  onChange,
  errors,
}) => {
  const selectedOptions = options.filter((e) => values.includes(e.value));

  return (
    <EuiComboBox
      fullWidth
      singleSelection={!isMultiValued}
      selectedOptions={selectedOptions}
      onChange={(items) => onChange(items.map((e) => e.value))}
      options={options}
      isInvalid={!!errors}
    />
  );
};

const CustomSegmenterConfigRow = ({
  isMultiValued,
  type,
  values,
  onChange,
  errors,
}) => {
  const [rawInput, setRawInput] = useState(values.join("\n"));

  const parseValues = useMemo(
    () => (value) => {
      // If multi-valued, split by new line. Else, place text in an array.
      const values = isMultiValued ? (value || "").split("\n") : [value];
      // Convert each item to the correct type
      const parsedValues = values.reduce((acc, e) => {
        try {
          const parsed = parseSegmenterValue(e.trim(), type);
          return [...acc, parsed];
        } catch (e) {
          return acc;
        }
      }, []);
      return parsedValues;
    },
    [isMultiValued, type]
  );

  useEffect(() => {
    // Only reset local state for raw input if `values` do not match the parsed input.
    // This can happen when `values` is updated by segment templates.
    // In all other scenarios, `values` will be updated when the raw input changes.
    if (!isEqual(parseValues(rawInput), values)) {
      setRawInput(values.join("\n"));
    }
  }, [values, parseValues, rawInput]);

  const onChangeText = (value) => {
    setRawInput(value);
    // Parse and save the input to the experiment segment.
    onChange(parseValues(value));
  };

  return isMultiValued ? (
    <EuiTextArea
      fullWidth
      compressed
      placeholder="Enter one value on each line"
      value={rawInput}
      onChange={(e) => onChangeText(e.target.value)}
      isInvalid={!!errors}
      resize="vertical"
    />
  ) : (
    <EuiFieldText
      fullWidth
      placeholder="Enter value"
      value={rawInput}
      onChange={(e) => onChangeText(e.target.value)}
      isInvalid={!!errors}
    />
  );
};

export const SegmenterConfigRow = ({
  name,
  type,
  isRequired,
  isMultiValued,
  description,
  options,
  values,
  onChange,
  errors,
}) => {
  const formattedName = `${name}${isRequired ? " *" : ""}`;

  // return only the first error since all the errors are of the form: Array elements must all be of type: X
  if (errors !== undefined) {
    errors = Object.values(errors)[0];
  }

  return (
    <EuiFlexItem grow={1}>
      <EuiFormRow
        fullWidth
        label={
          description.length > 0 ? (
            <FormLabelWithToolTip
              label={`${formattedName}`}
              content={`${description}`}
            />
          ) : (
            formattedName
          )
        }
        isInvalid={!!errors}
        error={errors}
        display="row">
        {options.length > 0 ? (
          <SelectableSegmenterConfigRow
            isMultiValued={isMultiValued}
            options={options}
            values={values}
            onChange={onChange}
            errors={errors}
          />
        ) : (
          <CustomSegmenterConfigRow
            isMultiValued={isMultiValued}
            type={type}
            values={values}
            onChange={onChange}
            errors={errors}
          />
        )}
      </EuiFormRow>
    </EuiFlexItem>
  );
};
