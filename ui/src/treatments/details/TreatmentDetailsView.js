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
import { useXpApi } from "hooks/useXpApi";
import { TreatmentConfigView } from "treatments/details/config/TreatmentConfigView";
import { TreatmentActions } from "treatments/details/TreatmentActions";
import EditTreatmentView from "treatments/edit/EditTreatmentView";
import ListTreatmentHistoryView from "treatments/history/ListTreatmentHistoryView";

const TreatmentDetailsView = ({ projectId, treatmentId, ...props }) => {
  const [{ data, isLoaded, error }, fetchTreatmentDetails] = useXpApi(
    `/projects/${projectId}/treatments/${treatmentId}`
  );

  const [{ data: historyData }, fetchTreatmentHistory] = useXpApi(
    `/projects/${projectId}/treatments/${treatmentId}/history`,
    {},
    { paging: { total: 0 } }
  );

  const onTreatmentChange = useCallback(() => {
    fetchTreatmentDetails();
    fetchTreatmentHistory();
  }, [fetchTreatmentDetails, fetchTreatmentHistory]);

  useEffect(() => {
    if ((props.location.state || {}).refresh) {
      onTreatmentChange();
    }
  }, [onTreatmentChange, props.location.state]);

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
                    <PageTitle title={data.data.name} />
                  </EuiPageHeaderSection>
                </EuiPageHeader>
                <TreatmentActions
                  onEdit={() => props.navigate("./edit")}
                  onDeleteSuccess={() => props.navigate("../")}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={getActions(data.data)}
                      selectedTab={props["*"]}
                      {...props}
                    />
                  )}
                </TreatmentActions>
                <EuiSpacer size="xl" />
              </Fragment>
            ) : (
              <EuiSpacer />
            )}
            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <TreatmentConfigView path="details" treatment={data.data} />
              <ListTreatmentHistoryView path="history" treatment={data.data} />

              <EditTreatmentView path="edit" treatmentSpec={data.data} />
            </Router>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default TreatmentDetailsView;
