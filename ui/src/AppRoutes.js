import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { useConfig } from "config";
import ExperimentsLandingPage from "experiments/ExperimentsLandingPage";
import Home from "Home";

const App = () => {
  const { appConfig } = useConfig();

  return (
    <Routes>
      {/* We need this redirection because XP is not a recognized MLP app and so PrivateLayout 
      should not redirect to /xp directly on project change. */}
      <Route path="projects/:projectId/*" element={<Navigate to={appConfig.homepage} replace={true} />} />
      {/* ALL ROUTES */}
      <Route path={appConfig.homepage}>
        <Route index element={<Home />} />
        <Route path="projects/:projectId">
          <Route index element={<Navigate to="experiments" replace={true} />} />
          <Route path="experiments/*" element={<ExperimentsLandingPage />} />
        </Route>
      </Route>
      {/* DEFAULT */}
      <Route path="*" element={<Navigate to="/pages/404" replace={true} />} />
    </Routes>
  );
};

export default App;
