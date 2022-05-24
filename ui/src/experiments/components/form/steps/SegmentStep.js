import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@gojek/mlp-ui";

import { SegmentSelectionPanel } from "experiments/components/form/components/segment_config/SegmentSelectionPanel";
import { SegmenterContextProvider } from "providers/segmenters/context";

export const SegmentStep = ({ projectId, isEdit }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <SegmenterContextProvider projectId={projectId}>
          <SegmentSelectionPanel
            isEdit={isEdit}
            projectId={projectId}
            segment={data.segment}
            segmentTemplate={data.segment_template}
            onChange={onChange}
            errors={errors}
          />
        </SegmenterContextProvider>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
