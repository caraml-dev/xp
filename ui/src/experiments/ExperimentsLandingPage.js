import React from "react";

import { Redirect, Router, useLocation } from "@reach/router";

import LandingView from "components/page/LandingView";
import { ProjectContextProvider } from "providers/project/context";

import CreateSegmentView from "../segments/create/CreateSegmentView";
import SegmentDetailsView from "../segments/details/SegmentDetailsView";
import SegmentHistoryDetailsView from "../segments/history/details/SegmentHistoryDetailsView";
import ListSegmentsView from "../segments/list/ListSegmentsView";
import CreateSettingsView from "../settings/create/CreateSettingsView";
import SettingsDetailsView from "../settings/details/SettingsDetailsView";
import CreateTreatmentView from "../treatments/create/CreateTreatmentView";
import TreatmentDetailsView from "../treatments/details/TreatmentDetailsView";
import TreatmentHistoryDetailsView from "../treatments/history/details/TreatmentHistoryDetailsView";
import ListTreatmentsView from "../treatments/list/ListTreatmentsView";
import CreateExperimentView from "./create/CreateExperimentView";
import ExperimentDetailsView from "./details/ExperimentDetailsView";
import ExperimentHistoryDetailsView from "./history/details/ExperimentHistoryDetailsView";
import ListExperimentsView from "./list/ListExperimentsView";

const ExperimentsLandingPage = ({ projectId, ...props }) => {
  /* Application Routes should be defined here, as ExperimentsLandingPage component is
     being exposed for use via MFE architecture. */
  return (
    <ProjectContextProvider>
      <Router>
        <CreateSettingsView path="/settings/create" />
        <SettingsDetailsView path="/settings/*" />

        <TreatmentHistoryDetailsView path="/treatments/:treatmentId/history/:version" />
        <LandingView
          Component={ListTreatmentsView}
          name="Treatments"
          projectId={projectId}
          path="/treatments"
        />
        <CreateTreatmentView path="/treatments/create" />
        <TreatmentDetailsView path="/treatments/:treatmentId/*" />

        <SegmentHistoryDetailsView path="/segments/:segmentId/history/:version" />
        <LandingView
          Component={ListSegmentsView}
          name="Segments"
          projectId={projectId}
          path="/segments"
        />
        <CreateSegmentView path="/segments/create" />
        <SegmentDetailsView path="/segments/:segmentId/*" />

        <CreateExperimentView path="/create" />
        <ExperimentDetailsView path="/:experimentId/*" />

        <ExperimentHistoryDetailsView path="/:experimentId/history/:version" />

        <LandingView
          Component={ListExperimentsView}
          name="Experiments"
          projectId={projectId}
          path="/"
        />

        {/* /experiments is the list view as well as a prefix to the other views which are registered without it;
        This redirect ensures that navigation from other views with /experiments prefix will not cause concatenation 
        which results in incorrect /experiments/experiments prefix.
         */}
        <Redirect
          from="/experiments/*"
          to={useLocation().pathname.replace(
            "/experiments/experiments",
            "/experiments"
          )}
          noThrow
        />
        <Redirect from="any" to="/error/404" default noThrow />
      </Router>
    </ProjectContextProvider>
  );
};

export default ExperimentsLandingPage;
