import React, { Fragment, useCallback, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";
import { Navigate, Route, Routes, useLocation, useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { useXpApi } from "hooks/useXpApi";
import { TreatmentConfigView } from "treatments/details/config/TreatmentConfigView";
import EditTreatmentView from "treatments/edit/EditTreatmentView";
import ListTreatmentHistoryView from "treatments/history/ListTreatmentHistoryView";

import { TreatmentActions } from "./TreatmentActions";
import { useConfig } from "config";

const TreatmentDetailsView = () => {
  const { projectId, treatmentId, "*": section } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

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
    if ((location.state || {}).refresh) {
      onTreatmentChange();
    }
  }, [onTreatmentChange, location.state]);

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
          {!(section === "edit") && (
            <Fragment>
              <EuiPageTemplate.Header
                bottomBorder={false}
                pageTitle={<PageTitle title={data.data.name} />}
              >
                <TreatmentActions
                  onEdit={() => navigate("./edit")}
                  onDeleteSuccess={() => navigate("../")}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={getActions(data.data)}
                      selectedTab={section}
                    />
                  )}
                </TreatmentActions>
              </EuiPageTemplate.Header>
            </Fragment>
          )}

          <Routes>
            <Route index element={<Navigate to="details" replace={true} />} />
            <Route path="details" element={<TreatmentConfigView treatment={data.data} />} />
            <Route path="history" element={<ListTreatmentHistoryView treatment={data.data} />} />
            <Route path="edit" element={<EditTreatmentView treatmentSpec={data.data} />} />
          </Routes>
        </Fragment>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default TreatmentDetailsView;
