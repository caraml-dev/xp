import React, { Fragment, useCallback, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation } from "@gojek/mlp-ui";
import { Redirect, Router } from "@reach/router";

import { PageTitle } from "components/page/PageTitle";
import { useXpApi } from "hooks/useXpApi";
import { SegmentConfigView } from "segments/details/config/SegmentConfigView";
import EditSegmentView from "segments/edit/EditSegmentView";
import ListSegmentHistoryView from "segments/history/ListSegmentHistoryView";

import { SegmentActions } from "./SegmentActions";
import { useConfig } from "../../config";

const SegmentDetailsView = ({ projectId, segmentId, ...props }) => {
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
    if ((props.location.state || {}).refresh) {
      onSegmentChange();
    }
  }, [onSegmentChange, props.location.state]);

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
                  pageTitle={<PageTitle title={data.data.name} />}
                >
                  <SegmentActions
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
                  </SegmentActions>
                </EuiPageTemplate.Header>
              </Fragment>
            )}

            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <SegmentConfigView path="details" segment={data.data} />
              <ListSegmentHistoryView path="history" segment={data.data} />
              <EditSegmentView path="edit" segmentSpec={data.data} />
            </Router>
          </Fragment>
        )}
    </EuiPageTemplate>
  );
};

export default SegmentDetailsView;
