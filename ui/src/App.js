import React from "react";

import {
  AuthProvider,
  Empty,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  PrivateRoute,
  Toast,
} from "@gojek/mlp-ui";
import { Redirect, Router } from "@reach/router";

import { useConfig } from "config";
import ExperimentsLandingPage from "experiments/ExperimentsLandingPage";
import Home from "Home";
import { PrivateLayout } from "PrivateLayout";

const App = () => {
  const { apiConfig, appConfig, authConfig } = useConfig();
  return (
    <ErrorBoundary>
      <MlpApiContextProvider
        mlpApiUrl={apiConfig.mlpApiUrl}
        timeout={apiConfig.apiTimeout}>
        <AuthProvider clientId={authConfig.oauthClientId}>
          <Router role="group">
            <Login path="/login" />

            <Redirect from="/" to={appConfig.homepage} noThrow />

            <Redirect
              from={`${appConfig.homepage}/projects/:projectId`}
              to={`${appConfig.homepage}/projects/:projectId/experiments`}
              noThrow
            />

            {/* HOME */}
            <PrivateRoute
              path={appConfig.homepage}
              render={PrivateLayout(Home)}
            />

            {/* EXPERIMENTS */}
            <PrivateRoute
              path={`${appConfig.homepage}/projects/:projectId/experiments/*`}
              render={PrivateLayout(ExperimentsLandingPage)}
            />

            {/* DEFAULT */}
            <Empty default />
          </Router>
          <Toast />
        </AuthProvider>
      </MlpApiContextProvider>
    </ErrorBoundary>
  );
};

export default App;
