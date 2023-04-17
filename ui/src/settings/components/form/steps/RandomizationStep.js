import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { SingleFieldConfigPanel } from "settings/components/form/components/config_section/SingleFieldConfigPanel";

export const RandomizationStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <SingleFieldConfigPanel
          toolTipLabel="Randomization Key *"
          toolTipContent="The name of the request field to be used for randomization. Eg: session_id"
          textValue={data.randomization_key}
          textPlaceHolder="Enter Randomization Key"
          onChange={onChange("randomization_key")}
          errors={errors?.randomization_key}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
