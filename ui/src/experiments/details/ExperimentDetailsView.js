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
import { Redirect, Router } from "@reach/router";

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

const ExperimentDetailsView = ({ projectId, experimentId, ...props }) => {
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
    if ((props.location.state || {}).refresh) {
      onExperimentChange();
    }
  }, [onExperimentChange, props.location.state]);

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
                  onEdit={() => props.navigate("./edit")}
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
          <Router primary={false}>
            <Redirect from="/" to="details" noThrow />
            <ExperimentConfigView path="details" experiment={data.data} />
            <ListExperimentHistoryView path="history" experiment={data.data} />
            <EditExperimentView path="edit" experimentSpec={data.data} />
          </Router>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate >
  );
};

export default ExperimentDetailsView;
