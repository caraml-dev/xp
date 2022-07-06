import React, { useContext } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@gojek/mlp-ui";

import { SegmentSelectionPanel } from "experiments/components/form/components/segment_config/SegmentSelectionPanel";
import { ConfigPanel } from "segments/components/form/components/config_section/ConfigPanel";

export const ConfigurationStep = ({ projectId, isEdit }) => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={true}>
        <ConfigPanel
          name={data.name}
          onChange={onChange}
          errors={errors}
          isEdit={!!data.id}
        />
        <EuiSpacer />
        <SegmentSelectionPanel
          isEdit={isEdit}
          projectId={projectId}
          segment={data.segment}
          segmentTemplate={data.segment_template}
          onChange={onChange}
          errors={errors}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
