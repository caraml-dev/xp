import React, { Fragment, useCallback, useEffect } from "react";

import {
  EuiCallOut,
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiSpacer,
  EuiTextAlign,
  EuiPageTemplate,
} from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";
import { Navigate, Router, Route, useLocation, useNavigate, useParams } from "react-router-dom";

import { VersionBadge } from "components/version_badge/VersionBadge";
import { StatusBadge } from "components/status_badge/StatusBadge";
import { PageTitle } from "components/page/PageTitle";
import EditExperimentView from "experiments/edit/EditExperimentView";
import ListExperimentHistoryView from "experiments/history/ListExperimentHistoryView";
import { useXpApi } from "hooks/useXpApi";
import { getExperimentStatus } from "services/experiment/ExperimentStatus";

import { ExperimentConfigView } from "./config/ExperimentConfigView";
import { ExperimentActions } from "./ExperimentActions";
import { useConfig } from "config";

const ExperimentBadges = ({ version, status }) => (
  <EuiFlexGroup wrap responsive={false} gutterSize="xs">
    <EuiFlexItem>
      <VersionBadge version={version} />
    </EuiFlexItem>
    <EuiFlexItem>
      <StatusBadge status={status} />
    </EuiFlexItem>
  </EuiFlexGroup>
);

const ExperimentDetailsView = () => {
  const { projectId, experimentId } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [{ data, isLoaded, error }, fetchExperimentDetails] = useXpApi(
    `/projects/${projectId}/experiments/${experimentId}`
  );

  const [{ data: historyData }, fetchExperimentHistory] = useXpApi(
    `/projects/${projectId}/experiments/${experimentId}/history`,
    {},
    { paging: { total: 0 } }
  );

  const onExperimentChange = useCallback(() => {
    fetchExperimentDetails();
    fetchExperimentHistory();
  }, [fetchExperimentDetails, fetchExperimentHistory]);

  useEffect(() => {
    if ((location.state || {}).refresh) {
      onExperimentChange();
    }
  }, [onExperimentChange, location.state]);

  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
    {
      id: "history",
      name: "History",
      disabled: !historyData.paging.total,
    },
  ];

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      {!isLoaded ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : error ? (
        <EuiCallOut
          title="Sorry, there was an error"
          color="danger"
          iconType="alert">
          <p>{error.message}</p>
        </EuiCallOut>
      ) : (
        <Fragment>
          {!(props["*"] === "edit") && (
            <Fragment>
              <EuiPageTemplate.Header
                bottomBorder={false}
                pageTitle={
                  <PageTitle
                    title={data.data.name}
                    postpend={
                      <ExperimentBadges status={getExperimentStatus(data.data)} version={data.data.version} />
                    }
                  />
                }
              >
                <ExperimentActions
                  onEdit={() => navigate("./edit")}
                  onActivateSuccess={onExperimentChange}
                  onDeactivateSuccess={onExperimentChange}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={getActions(data.data)}
                      selectedTab={props["*"]}
                      {...props}
                    />
                  )}
                </ExperimentActions>
              </EuiPageTemplate.Header>
            </Fragment>
          )}
          <Routes>
            <Route index element={<Navigate to="details" replace={true} />} />
            <Route path="details" element={<ExperimentConfigView experiment={data.data} />} />
            <Route path="history" element={<ListExperimentHistoryView experiment={data.data} />} />
            <Route path="edit" element={<EditExperimentView experimentSpec={data.data} />} />
          </Routes>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default ExperimentDetailsView;
