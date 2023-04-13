import React from "react";

import {
  EuiDatePicker,
  EuiDatePickerRange,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "@caraml-dev/ui-lib";
import moment from "moment";

import SuperSelectWithDescription from "components/form/select/SuperSelectWithDescription";
import { Panel } from "components/panel/Panel";
import { useConfig } from "config";

export const ScheduleConfigPanel = ({
  status,
  startTime,
  endTime,
  statusOptions,
  onChange,
  errors = {},
}) => {
  const { appConfig } = useConfig();
  const onChangeTime = (time_field) => (time) => {
    let utcTime = time;
    if (time.utcOffset !== 0) {
      // Browser would treat moment selected as local timezone
      // Hence, we'll need to force convert all timezones to UTC since we're
      // asking users to input UTC timestamps.
      utcTime = moment.utc(time.format(appConfig.datetime.formatNoTz));
    }

    onChange(time_field)(utcTime);
  };

  return (
    <Panel title="Schedule">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={1}>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Status *"
                  content="Specify the experiment status."
                />
              }
              isInvalid={!!errors.status}
              error={errors.status}
              display="row">
              <SuperSelectWithDescription
                fullWidth
                value={status}
                onChange={onChange("status")}
                options={statusOptions}
                hasDividers
                isInvalid={!!errors.status}
              />
            </EuiFormRow>
          </EuiFlexItem>
          <EuiFlexItem grow={2}>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label={`Duration (${appConfig.datetime.tz}) *`}
                  content="Specify the experiment start and end time in UTC."
                />
              }
              isInvalid={!!errors.start_time || !!errors.end_time}
              error={[errors.start_time, errors.end_time]}
              display="row">
              <EuiDatePickerRange
                fullWidth
                startDateControl={
                  <EuiDatePicker
                    selected={startTime}
                    onChange={onChangeTime("start_time")}
                    startDate={startTime}
                    endDate={endTime}
                    isInvalid={!!errors.start_time}
                    aria-label="Start Time"
                    showTimeSelect
                    utcOffset={appConfig.datetime.tzOffsetMinutes}
                  />
                }
                endDateControl={
                  <EuiDatePicker
                    selected={endTime}
                    onChange={onChangeTime("end_time")}
                    startDate={startTime}
                    endDate={endTime}
                    isInvalid={!!errors.end_time}
                    aria-label="End Time"
                    showTimeSelect
                    utcOffset={appConfig.datetime.tzOffsetMinutes}
                  />
                }
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </Panel>
  );
};
