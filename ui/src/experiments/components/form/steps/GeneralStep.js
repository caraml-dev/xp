import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@gojek/mlp-ui";

import { GeneralSettingsConfigPanel } from "experiments/components/form/components/general_config/GeneralSettingsConfigPanel";
import { ScheduleConfigPanel } from "experiments/components/form/components/general_config/ScheduleConfigPanel";
import { SwitchbackConfigPanel } from "experiments/components/form/components/general_config/SwitchbackConfigPanel";
import {
  experimentStatuses,
  experimentTiers,
  experimentTypes,
} from "experiments/components/typeOptions";

export const GeneralStep = ({ projectId }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  const experimentTypeOptions = experimentTypes.map((e) => ({
    ...e,
    inputDisplay: e.inputDisplay || e.label,
  }));

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <GeneralSettingsConfigPanel
          name={data.name}
          type={data.type}
          tier={data.tier}
          description={data.description}
          typeOptions={experimentTypeOptions}
          tierOptions={experimentTiers}
          isEdit={!!data.id}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
      <EuiFlexItem grow={true}>
        <ScheduleConfigPanel
          status={data.status}
          startTime={data.start_time}
          endTime={data.end_time}
          statusOptions={experimentStatuses}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
      {data.type === "Switchback" && (
        <EuiFlexItem grow={true}>
          <SwitchbackConfigPanel
            interval={data.interval}
            onChange={onChange}
            errors={errors}
          />
        </EuiFlexItem>
      )}
    </EuiFlexGroup>
  );
};
