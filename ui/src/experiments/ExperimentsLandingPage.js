import React from "react";

import { Navigate, Route, Routes, useLocation } from "react-router-dom";

import LandingView from "components/page/LandingView";
import { ConfigProvider } from "config";
import { ProjectContextProvider } from "providers/project/context";

import CreateExperimentView from "./create/CreateExperimentView";
import ExperimentDetailsView from "./details/ExperimentDetailsView";
import ExperimentHistoryDetailsView from "./history/details/ExperimentHistoryDetailsView";
import ListExperimentsView from "./list/ListExperimentsView";
import CreateSegmentView from "segments/create/CreateSegmentView";
import SegmentDetailsView from "segments/details/SegmentDetailsView";
import SegmentHistoryDetailsView from "segments/history/details/SegmentHistoryDetailsView";
import ListSegmentsView from "segments/list/ListSegmentsView";
import CreateSettingsView from "settings/create/CreateSettingsView";
import SettingsDetailsView from "settings/details/SettingsDetailsView";
import CreateTreatmentView from "treatments/create/CreateTreatmentView";
import TreatmentDetailsView from "treatments/details/TreatmentDetailsView";
import TreatmentHistoryDetailsView from "treatments/history/details/TreatmentHistoryDetailsView";
import ListTreatmentsView from "treatments/list/ListTreatmentsView";

const ExperimentsLandingPage = () => {
  const location = useLocation();
  /* Application Routes should be defined here, as ExperimentsLandingPage component is
     being exposed for use via MFE architecture. */
  return (
    <ConfigProvider>
      <ProjectContextProvider>
        <Routes>
          {/* SETTINGS */}
          <Route path="settings">
            <Route index path="*" element={<SettingsDetailsView />} />
            <Route path="create" element={<CreateSettingsView />} />
          </Route>

          {/* TREATMENTS */}
          <Route path="treatments">
            <Route index element={<LandingView Component={ListTreatmentsView} name="Treatments" />} />
            <Route path="create" element={<CreateTreatmentView />} />
            <Route path=":treatmentId">
              <Route path="history/:version" element={<TreatmentHistoryDetailsView />} />
              <Route index path="*" element={<TreatmentDetailsView />} />
            </Route>
          </Route>

          {/* SEGMENTS */}
          <Route path="segments">
            <Route index element={<LandingView Component={ListSegmentsView} name="Segments" />} />
            <Route path="create" element={<CreateSegmentView />} />
            <Route path=":segmentId">
              <Route path="history/:version" element={<SegmentHistoryDetailsView />} />
              <Route index path="*" element={<SegmentDetailsView />} />
            </Route>
          </Route>

          {/* EXPERIMENTS */}
          <Route index element={<LandingView Component={ListExperimentsView} name="Experiments" />} />
          <Route path="create" element={<CreateExperimentView />} />
          <Route path=":experimentId">
            <Route path="history/:version" element={<ExperimentHistoryDetailsView />} />
            <Route index path="*" element={<ExperimentDetailsView />} />
          </Route>

          {/* /experiments is the list view as well as a prefix to the other views which are registered without it;
        This redirect ensures that navigation from other views with /experiments prefix will not cause concatenation 
        which results in incorrect /experiments/experiments prefix.
         */}
          <Route
            path="experiments"
            element={<Navigate to={location.pathname.replace("/experiments/experiments", "/experiments")}
              replace={true} />} />
        </Routes>
      </ProjectContextProvider>
    </ConfigProvider>
  );
};

export default ExperimentsLandingPage;
