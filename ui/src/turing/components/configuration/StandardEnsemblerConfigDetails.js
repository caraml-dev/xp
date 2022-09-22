import React from "react";
import { ConfigProvider, useConfig } from "config";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "components/config_section/ConfigSectionPanel";
import { LinkedRoutesTable } from "../form/standard_ensembler/LinkedRoutesTable";
import { RouteNamePathConfigGroup } from "./standard_ensembler_config/RouteNamePathConfigGroup";
import { ProjectContextProvider } from "providers/project/context";

const StandardEnsemblerConfigDetailsComponent = ({ projectId, routes, routeNamePath }) => {
  const { appConfig: { routeNamePathPrefix } } = useConfig();

  return (
    <ConfigProvider>
      <EuiFlexGroup direction="row" wrap>
        <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
          <ConfigSectionPanel title="Route Selection" className="experimentSummaryPanel">
            <RouteNamePathConfigGroup routeNamePath={routeNamePath} />
          </ConfigSectionPanel>
        </EuiFlexItem>

        <EuiFlexItem grow={2} className="euiFlexItem--smallPanel">
          <ConfigSectionPanel title="Linked Routes" className="linkedRoutesPanel">
            <LinkedRoutesTable
              projectId={projectId}
              routes={routes}
              treatmentConfigRouteNamePath={routeNamePath.slice(routeNamePathPrefix.length)}
            />
          </ConfigSectionPanel>
        </EuiFlexItem>
      </EuiFlexGroup>
    </ConfigProvider>
  );
};

const StandardEnsemblerConfigDetails = (props) => {
  return (
    <ConfigProvider>
      <ProjectContextProvider>
        <StandardEnsemblerConfigDetailsComponent {...props} />
      </ProjectContextProvider>
    </ConfigProvider>
  )
};

export default StandardEnsemblerConfigDetails;
