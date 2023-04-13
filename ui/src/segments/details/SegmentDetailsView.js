import React, { Fragment, useCallback, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation } from "@caraml-dev/ui-lib";
import { Navigate, Route, Routes, useLocation, useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { useXpApi } from "hooks/useXpApi";
import { SegmentConfigView } from "segments/details/config/SegmentConfigView";
import EditSegmentView from "segments/edit/EditSegmentView";
import ListSegmentHistoryView from "segments/history/ListSegmentHistoryView";

import { SegmentActions } from "./SegmentActions";
import { useConfig } from "config";

const SegmentDetailsView = () => {
  const { projectId, segmentId, "*": section } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [{ data, isLoaded, error }, fetchSegmentDetails] = useXpApi(
    `/projects/${projectId}/segments/${segmentId}`
  );

  const [{ data: historyData }, fetchSegmentHistory] = useXpApi(
    `/projects/${projectId}/segments/${segmentId}/history`,
    {},
    { paging: { total: 0 } }
  );

  const onSegmentChange = useCallback(() => {
    fetchSegmentDetails();
    fetchSegmentHistory();
  }, [fetchSegmentDetails, fetchSegmentHistory]);

  useEffect(() => {
    if ((location.state || {}).refresh) {
      onSegmentChange();
    }
  }, [onSegmentChange, location.state]);

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
                <SegmentActions
                  onEdit={() => navigate("./edit")}
                  onDeleteSuccess={() => navigate("../")}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={getActions(data.data)}
                      selectedTab={section}
                    />
                  )}
                </SegmentActions>
              </EuiPageTemplate.Header>
            </Fragment>
          )}

          <Routes>
            <Route index element={<Navigate to="details" replace={true} />} />
            <Route path="details" element={<SegmentConfigView segment={data.data} />} />
            <Route path="history" element={<ListSegmentHistoryView segment={data.data} />} />
            <Route path="edit" element={<EditSegmentView segmentSpec={data.data} />} />
          </Routes>
        </Fragment>
      )}
    </EuiPageTemplate>
  );
};

export default SegmentDetailsView;
