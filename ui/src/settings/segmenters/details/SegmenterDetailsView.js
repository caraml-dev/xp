import React, { Fragment, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { Redirect, Router } from "@reach/router";

import { PageTitle } from "components/page/PageTitle";
import { StatusBadge } from "components/status_badge/StatusBadge";
import { useXpApi } from "hooks/useXpApi";
import { getSegmenterStatus } from "services/segmenter/SegmenterStatus";
import { SegmentersConfigView } from "settings/segmenters/details/config/SegmentersConfigView";
import { SegmenterActions } from "settings/segmenters/details/SegmenterActions";
import EditSegmenterView from "settings/segmenters/edit/EditSegmenterView";

const SegmenterDetailsView = ({ projectId, segmenterName, ...props }) => {
  const tabs = [
    {
      id: "details",
      name: "Configuration",
    },
  ];

  const [
    {
      data: { data: segmenter },
      isLoaded,
      error,
    },
  ] = useXpApi(`/projects/${projectId}/segmenters/${segmenterName}`, {}, []);

  // ../../segmenters is required, with pure .. it will end up with /segmenters/ and the tab routing is bugged
  useEffect(() => {
    !!segmenter &&
      replaceBreadcrumbs([
        { text: "Segmenters", href: "../../segmenters" },
        { text: segmenter.name },
        { text: "Configuration" },
      ]);
  }, [segmenter]);

  return (
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
                    icon="package"
                    title={segmenter.name}
                    postpend={
                      <StatusBadge status={getSegmenterStatus(segmenter)} />
                    }
                  />
                </EuiPageHeaderSection>
              </EuiPageHeader>
              <EuiSpacer size="m" />
              {props["*"] === "details" && (
                <SegmenterActions
                  onEdit={() => props.navigate("./edit")}
                  onDeleteSuccess={() => props.navigate("../")}
                  segmenter={segmenter}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={
                        segmenter.scope === "project"
                          ? getActions({
                            name: segmenterName,
                            projectId: projectId,
                          })
                          : null
                      }
                      selectedTab={props["*"]}
                      {...props}
                    />
                  )}
                </SegmenterActions>
              )}
              <EuiSpacer size="xl" />
            </Fragment>
          ) : (
            <EuiSpacer />
          )}

          <Router primary={false}>
            <Redirect from="/" to="details" noThrow />
            <SegmentersConfigView path="details" segmenter={segmenter} />
            <EditSegmenterView path="edit" segmenter={segmenter} />
          </Router>
        </Fragment>
      )}
    </EuiPageBody>
  );
};

export default SegmenterDetailsView;
