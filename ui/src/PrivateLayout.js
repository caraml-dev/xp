import "PrivateLayout.scss";

import React from "react";

import {
  ApplicationsContext,
  ApplicationsContextProvider,
  Header,
  PrivateRoute,
  ProjectsContextProvider,
} from "@gojek/mlp-ui";
import urlJoin from "proper-url-join";
import { Outlet, useNavigate } from "react-router-dom";

import { useConfig } from "config";

export const PrivateLayout = () => {
  const navigate = useNavigate();
  const { appConfig } = useConfig();
  return (
    <PrivateRoute>
      <ApplicationsContextProvider>
        <ProjectsContextProvider>
          <ApplicationsContext.Consumer>
            {({ currentApp }) => (
              <Header
                homepage={appConfig.homepage}
                onProjectSelect={pId =>
                  navigate(urlJoin(currentApp?.href, "projects", pId, "experiments"))
                }
                docLinks={appConfig.docsUrl}
              />)}
          </ApplicationsContext.Consumer>
          <Outlet />
        </ProjectsContextProvider>
      </ApplicationsContextProvider>
    </PrivateRoute>
  );
};
