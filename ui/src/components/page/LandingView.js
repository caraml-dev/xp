import React, { Fragment, useContext } from "react";

import {
  EuiLink,
  EuiLoadingChart,
  EuiSpacer,
  EuiText,
  EuiTextAlign,
  EuiPageTemplate
} from "@elastic/eui";

import ProjectContext from "providers/project/context";
import { useNavigate, useParams } from "react-router-dom";

const LandingView = ({ Component, name }) => {
  const navigate = useNavigate();
  const { projectId } = useParams();
  const { isProjectOnboarded, isLoaded } = useContext(ProjectContext);

  return (
    <>
      {!isLoaded ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : !isProjectOnboarded(projectId) ? (
        <EuiPageTemplate panelled={false}>
          <EuiPageTemplate.EmptyPrompt
            title={
              <EuiText>
                Welcome to {name}.{"\n\n"}
              </EuiText>
            }
            body={
              <Fragment>
                <EuiSpacer />
                <EuiText>
                  This project has not been set up. Get started{" "}
                  <EuiLink onClick={() => navigate("./settings/create")}>
                    here
                  </EuiLink>
                  .
                </EuiText>
              </Fragment>
            }
          />
        </EuiPageTemplate>
      ) : (
        <Component projectId={projectId} />
      )}
    </>
  );
};

export default LandingView;
