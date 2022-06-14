import React, { useContext, useEffect, useRef } from "react";

import { EuiCallOut, EuiFlexItem, EuiLoadingChart } from "@elastic/eui";
import { OverlayMask, get, useOnChangeHandler } from "@gojek/mlp-ui";

import { Panel } from "components/panel/Panel";
import { ConfigProvider } from "config";
import ProjectContext, {
  ProjectContextProvider,
} from "providers/project/context";
import { SettingsContextProvider } from "providers/settings/context";

import { VariablesConfigPanel } from "./variables_config/VariablesConfigPanel";

const EditExperimentEngineConfigComponent = ({
  projectId,
  config = {},
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { isProjectOnboarded, isLoaded } = useContext(ProjectContext);
  const overlayRef = useRef();

  useEffect(() => {
    // Set project id if not already set. This is only needed because the CustomExperimentManager
    // methods cannot take in a project ID as an argument and expect all the required data to
    // be stored in the config (which is ultimately saved to the Turing DB). In the future,
    // we could refactor this workflow such that the `project_id` value will not have to be
    // saved to the config at all - instead, it should use the project ID of the Turing router.
    if (!config.project_id) {
      onChange("project_id")(parseInt(projectId));
    }
  }, [projectId, config.project_id, onChange]);

  return (
    <EuiFlexItem grow={false}>
      {isLoaded ? (
        isProjectOnboarded(projectId) ? (
          <SettingsContextProvider projectId={projectId}>
            <VariablesConfigPanel
              projectId={projectId}
              variables={config.variables}
              onChangeHandler={onChange("variables")}
              errors={get(errors, "variables")}
            />
          </SettingsContextProvider>
        ) : (
          <Panel title="Configuration">
            <EuiCallOut
              title="Project not onboarded to Experiments"
              color="danger"
              iconType="alert">
              <p>
                {
                  "Please complete onboarding to Turing experiments to configure the router."
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

const EditExperimentEngineConfig = (props) => (
  <ConfigProvider>
    <ProjectContextProvider>
      <EditExperimentEngineConfigComponent {...props} />
    </ProjectContextProvider>
  </ConfigProvider>
);

export default EditExperimentEngineConfig;
