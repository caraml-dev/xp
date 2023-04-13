import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { SingleFieldConfigPanel } from "settings/components/form/components/config_section/SingleFieldConfigPanel";

export const ExternalValidationStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <SingleFieldConfigPanel
          toolTipLabel="Url"
          toolTipContent="External Url to be used for validation"
          textValue={data?.validation_url}
          textPlaceHolder="Enter External Validation Url"
          onChange={onChange("validation_url")}
          errors={errors?.validation_url}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
