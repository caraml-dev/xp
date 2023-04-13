import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { SegmenterPanel } from "settings/components/form/components/segmenter_section/SegmenterPanel";

export const SegmentersStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <SegmenterPanel
          segmenters={data.segmenters}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
