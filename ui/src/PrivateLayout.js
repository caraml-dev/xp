import "PrivateLayout.scss";

import React from "react";

import {
  ApplicationsContextProvider,
  CurrentProjectContextProvider,
  Header,
  ProjectsContextProvider,
} from "@gojek/mlp-ui";
import { navigate } from "@reach/router";

import { appConfig } from "config";

export const PrivateLayout = (Component) => {
  return (props) => (
    <ApplicationsContextProvider>
      <ProjectsContextProvider>
        <CurrentProjectContextProvider {...props}>
          <Header
            homeUrl={appConfig.homepage}
            appIcon={appConfig.appIcon}
            onProjectSelect={(projectId) =>
              navigate(
                `${appConfig.homepage}/projects/${projectId}/experiments`
              )
            }
            docLinks={appConfig.docsUrl}
          />
          <div className="main-component-layout">
            <Component {...props} />
          </div>
        </CurrentProjectContextProvider>
      </ProjectsContextProvider>
    </ApplicationsContextProvider>
  );
};
