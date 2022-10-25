import React from "react";

import {
  AuthProvider,
  Page404,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  Toast,
} from "@gojek/mlp-ui";
import { Route, Routes } from "react-router-dom";

import { useConfig } from "config";
import { PrivateLayout } from "PrivateLayout";
import { EuiProvider } from "@elastic/eui";
import AppRoutes from "AppRoutes";

const App = () => {
  const { apiConfig, authConfig } = useConfig();
  return (
    <EuiProvider>
      <ErrorBoundary>
        <MlpApiContextProvider
          mlpApiUrl={apiConfig.mlpApiUrl}
          timeout={apiConfig.apiTimeout}>
          <AuthProvider clientId={authConfig.oauthClientId}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route element={<PrivateLayout />}>
                <Route path="/*" element={<AppRoutes />} />
              </Route>
              <Route path="/pages/404" element={<Page404 />} />
            </Routes>
            <Toast />
          </AuthProvider>
        </MlpApiContextProvider>
      </ErrorBoundary>
    </EuiProvider>
  );
};

export default App;
