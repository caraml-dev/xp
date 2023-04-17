import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { OptionsPanel } from "settings/segmenters/components/form/components/OptionsPanel";

export const OptionsStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <OptionsPanel
          options={data.options}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
