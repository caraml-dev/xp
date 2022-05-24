import React, { useContext } from "react";

import {
  EuiLink,
  EuiLoadingChart,
  EuiPage,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiTextAlign,
} from "@elastic/eui";

import ProjectContext from "providers/project/context";

const LandingView = ({ Component, name, projectId, ...props }) => {
  const { isProjectOnboarded, isLoaded } = useContext(ProjectContext);

  return (
    <>
      {!isLoaded ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : !isProjectOnboarded(projectId) ? (
        <EuiPage>
          <EuiPanel>
            <EuiText>
              Welcome to {name}.{"\n\n"}
            </EuiText>
            <EuiSpacer />
            <EuiText>
              This project has not been set up. Get started{" "}
              <EuiLink onClick={() => props.navigate("./settings/create")}>
                here
              </EuiLink>
              .
            </EuiText>
          </EuiPanel>
        </EuiPage>
      ) : (
        <Component projectId={projectId} {...props} />
      )}
    </>
  );
};

export default LandingView;
