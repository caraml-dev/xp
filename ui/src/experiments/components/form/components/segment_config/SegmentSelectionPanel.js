import React, { useContext } from "react";

import {
  EuiFlexGroup,
  EuiLoadingChart,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";

import { Panel } from "components/panel/Panel";
import { SegmentConfigPanel } from "experiments/components/form/components/segment_config/SegmentConfigPanel";
import SegmentsContext from "providers/segment/context";

export const SegmentSelectionPanel = ({
  projectId,
  segment,
  segmentTemplate,
  onChange,
  errors = [],
}) => {
  const { isLoaded, segments } = useContext(SegmentsContext);

  const segmentSelectionOptions = segments.map((segment) => {
    return {
      label: segment.name,
      id: segment.id,
    };
  });

  return isLoaded ? (
    <Panel title="Segment Configuration">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="s">
        <SegmentConfigPanel
          projectId={projectId}
          segment={segment}
          segmentTemplate={segmentTemplate}
          onChange={onChange}
          segmentSelectionOptions={segmentSelectionOptions}
          errors={errors}
        />
      </EuiFlexGroup>
    </Panel>
  ) : (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  );
};
