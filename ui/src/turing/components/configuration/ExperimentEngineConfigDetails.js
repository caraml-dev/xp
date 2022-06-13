import React from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";

import { ConfigSectionPanel } from "components/config_section/ConfigSectionPanel";
import { ConfigProvider } from "config";

import { ExperimentsConfigGroup } from "./experiments_config/ExperimentsConfigGroup";
import { VariablesConfigGroup } from "./variables_config/VariablesConfigGroup";

const ExperimentEngineConfigDetails = ({ projectId, config }) => (
  <ConfigProvider>
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title="Summary" className="experimentSummaryPanel">
          <ExperimentsConfigGroup projectId={projectId} />
        </ConfigSectionPanel>
      </EuiFlexItem>

      <EuiFlexItem grow={2} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel
          title="Variables"
          className="experimentVariablesPanel">
          <VariablesConfigGroup variables={config.variables} />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  </ConfigProvider>
);

export default ExperimentEngineConfigDetails;
