import React, { useContext, useRef } from "react";

import { EuiCallOut, EuiFlexItem, EuiLoadingChart, EuiHorizontalRule } from "@elastic/eui";
import { OverlayMask } from "@gojek/mlp-ui";

import { Panel } from "components/panel/Panel";
import { ConfigProvider } from "config";
import ProjectContext, {
  ProjectContextProvider,
} from "providers/project/context";
import { SettingsContextProvider } from "providers/settings/context";
import { AffectedRoutesTable } from "./AffectedRoutesTable";
import { RouteNamePathRow } from "./RouteNamePathRow";

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
    <EuiFlexItem grow={false}>
      {isLoaded ? (
        isProjectOnboarded(projectId) ? (
          <SettingsContextProvider projectId={projectId}>
            <Panel title={"Route Selection"}>
              <RouteNamePathRow
                routeNamePath={routeNamePath}
                onChange={onChange}
                errors={errors}
              />

              <EuiHorizontalRule />

              <AffectedRoutesTable
                projectId={projectId}
                routes={routes}
                routeNamePath={routeNamePath}
                onChange={onChange}
                errors={errors}
              />
            </Panel>
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
    </EuiFlexItem>
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
