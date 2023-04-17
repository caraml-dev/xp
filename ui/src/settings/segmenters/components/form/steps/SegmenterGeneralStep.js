import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { SegmenterGeneralPanel } from "settings/segmenters/components/form/components/SegmenterGeneralPanel";

export const SegmenterGeneralStep = ({ isEdit }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <SegmenterGeneralPanel
          name={data.name}
          type={data.type}
          description={data.description}
          required={data.required}
          multiValued={data.multi_valued}
          onChange={onChange}
          errors={errors}
          isEdit={isEdit}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
