import React, { useContext, useRef } from "react";

import { EuiCallOut, EuiFlexItem, EuiLoadingChart, EuiFlexGroup } from "@elastic/eui";
import { OverlayMask } from "@gojek/mlp-ui";

import { Panel } from "components/panel/Panel";
import { ConfigProvider } from "config";
import ProjectContext, {
  ProjectContextProvider,
} from "providers/project/context";
import { SettingsContextProvider } from "providers/settings/context";
import { AffectedRoutesListPanel } from "./AffectedRoutesListPanel";
import { RouteNamePathPanel } from "./RouteNamePathPanel";

const EditStandardEnsemblerConfigComponent = ({
  projectId,
  routes,
  routeNamePath = "",
  onChange,
  errors,
}) => {
  const { isProjectOnboarded, isLoaded } = useContext(ProjectContext);
  const overlayRef = useRef();

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      {isLoaded ? (
        isProjectOnboarded(projectId) ? (
          <SettingsContextProvider projectId={projectId}>
            <EuiFlexItem>
              <RouteNamePathPanel
                projectId={projectId}
                routeNamePath={routeNamePath}
                onChange={onChange}
                errors={errors}
              />
            </EuiFlexItem>

            <EuiFlexItem>
              <AffectedRoutesListPanel
                projectId={projectId}
                routes={routes}
                routeNamePath={routeNamePath}
              />
            </EuiFlexItem>
          </SettingsContextProvider>
        ) : (
          <Panel title="Configuration">
            <EuiCallOut
              title="Project not onboarded to Experiments"
              color="danger"
              iconType="alert">
              <p>
                {
                  "Please complete onboarding to Turing experiments to configure the standard ensembler."
                }
              </p>
            </EuiCallOut>
          </Panel>
        )
      ) : (
        <div ref={overlayRef}>
          <OverlayMask parentRef={overlayRef} opacity={0.4}>
            <EuiLoadingChart size="xl" mono />
          </OverlayMask>
        </div>
      )}
    </EuiFlexGroup>
  );
};

const EditStandardEnsemblerConfig = (props) => {

  return (
    <ConfigProvider>
      <ProjectContextProvider>
        <EditStandardEnsemblerConfigComponent {...props} />
      </ProjectContextProvider>
    </ConfigProvider>
  )
};

export default EditStandardEnsemblerConfig;
