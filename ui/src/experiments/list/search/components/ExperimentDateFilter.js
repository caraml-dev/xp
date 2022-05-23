import React from "react";

import { EuiDatePicker, EuiDatePickerRange, EuiFormRow } from "@elastic/eui";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

import { appConfig } from "config";

const ExperimentDateFilter = ({ startTime, endTime, onChange, errors }) => {
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
            popperPlacement="bottom-start"
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
            popperPlacement="bottom-end"
          />
        }
      />
    </EuiFormRow>
  );
};

export default ExperimentDateFilter;
