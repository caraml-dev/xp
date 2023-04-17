import "PrivateLayout.scss";

import React from "react";

import {
  ApplicationsContext,
  ApplicationsContextProvider,
  Header,
  PrivateRoute,
  ProjectsContextProvider,
} from "@caraml-dev/ui-lib";
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
            {() => (
              <Header
                onProjectSelect={pId =>
                  navigate(urlJoin("projects", pId))
                }
                docLinks={appConfig.docsUrl}
              />
            )}
          </ApplicationsContext.Consumer>
          <Outlet />
        </ProjectsContextProvider>
      </ApplicationsContextProvider>
    </PrivateRoute>
  );
};
