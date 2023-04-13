import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { FormContext, FormValidationContext, get } from "@caraml-dev/ui-lib";

import { ConstraintsPanel } from "settings/segmenters/components/form/components/ConstraintsPanel";

export const ConstraintsStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <ConstraintsPanel
          constraints={data.constraints}
          onChangeHandler={onChangeHandler}
          errors={get(errors, "constraints")}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
