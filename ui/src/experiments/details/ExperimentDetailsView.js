import React, { Fragment, useCallback, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";
import { Redirect, Router } from "@reach/router";

import { PageTitle } from "components/page/PageTitle";
import { StatusBadge } from "components/status_badge/StatusBadge";
import EditExperimentView from "experiments/edit/EditExperimentView";
import ListExperimentHistoryView from "experiments/history/ListExperimentHistoryView";
import { useXpApi } from "hooks/useXpApi";
import { getExperimentStatus } from "services/experiment/ExperimentStatus";

import { ExperimentConfigView } from "./config/ExperimentConfigView";
import { ExperimentActions } from "./ExperimentActions";

const ExperimentDetailsView = ({ projectId, experimentId, ...props }) => {
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
    <EuiPage>
      <EuiPageBody>
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
            {!(props["*"] === "edit") ? (
              <Fragment>
                <EuiPageHeader>
                  <EuiPageHeaderSection>
                    <PageTitle
                      title={data.data.name}
                      postpend={
                        <StatusBadge status={getExperimentStatus(data.data)} />
                      }
                    />
                  </EuiPageHeaderSection>
                </EuiPageHeader>
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
                <EuiSpacer size="xl" />
              </Fragment>
            ) : (
              <EuiSpacer />
            )}

            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <ExperimentConfigView path="details" experiment={data.data} />

              <ListExperimentHistoryView
                path="history"
                experiment={data.data}
              />

              <EditExperimentView path="edit" experimentSpec={data.data} />
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default ExperimentDetailsView;
