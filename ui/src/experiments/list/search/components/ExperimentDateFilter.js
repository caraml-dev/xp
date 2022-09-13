import React from "react";

import { EuiDatePicker, EuiDatePickerRange, EuiFormRow } from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

import { useConfig } from "config";

const ExperimentDateFilter = ({ startTime, endTime, onChange, errors }) => {
  const { appConfig } = useConfig();
  return (
    <EuiFormRow
      fullWidth
      label={
        <FormLabelWithToolTip
          label="Duration"
          content="Any experiment whose duration partially overlaps with the specified range will be matched."
        />
      }
      isInvalid={!!errors?.start_time || !!errors?.end_time}
      error={[errors?.start_time, errors?.end_time]}>
      <EuiDatePickerRange
        startDateControl={
          <EuiDatePicker
            selected={startTime}
            onChange={onChange("start_time")}
            startDate={startTime}
            endDate={endTime}
            aria-label="Start Time"
            utcOffset={appConfig.datetime.tzOffsetMinutes}
            popoverPlacement="downLeft"
          />
        }
        endDateControl={
          <EuiDatePicker
            selected={endTime}
            onChange={onChange("end_time")}
            startDate={startTime}
            endDate={endTime}
            aria-label="End Time"
            utcOffset={appConfig.datetime.tzOffsetMinutes}
            popoverPlacement="downRight"
          />
        }
      />
    </EuiFormRow>
  );
};

export default ExperimentDateFilter;
