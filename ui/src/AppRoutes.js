import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { useConfig } from "config";
import ExperimentsLandingPage from "experiments/ExperimentsLandingPage";
import Home from "Home";

const App = () => {
  const { appConfig } = useConfig();
  return (
    <Routes>
      <Route path={appConfig.homepage} />
      <Route index element={<Home />} />
      <Route path="projects/:projectId">
        <Route index element={<Navigate to="experiments" replace={true} />} />
        <Route path="experiments/*" element={<ExperimentsLandingPage />} />
      </Route>
      {/* DEFAULT */}
      <Route path="*" element={<Navigate to="/pages/404" replace={true} />} />
    </Routes>
  );
};

export default App;
