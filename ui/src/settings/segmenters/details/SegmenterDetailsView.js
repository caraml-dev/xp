import React, { Fragment, useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiTextAlign,
  EuiPageTemplate,
} from "@elastic/eui";
import { PageNavigation, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { Navigate, Route, Routes, useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { StatusBadge } from "components/status_badge/StatusBadge";
import { useXpApi } from "hooks/useXpApi";
import { getSegmenterStatus } from "services/segmenter/SegmenterStatus";
import { SegmentersConfigView } from "settings/segmenters/details/config/SegmentersConfigView";
import { SegmenterActions } from "settings/segmenters/details/SegmenterActions";
import EditSegmenterView from "settings/segmenters/edit/EditSegmenterView";

const SegmenterDetailsView = () => {
  const { projectId, segmenterName, "*": section } = useParams();
  const navigate = useNavigate();

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
    <Fragment>
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
                pageTitle={
                  <PageTitle
                    icon="package"
                    title={segmenter.name}
                    postpend={
                      <StatusBadge status={getSegmenterStatus(segmenter)} />
                    }
                  />
                }
              >
                {section === "details" && (
                  <SegmenterActions
                    onEdit={() => navigate("./edit")}
                    onDeleteSuccess={() => navigate("../")}
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
                        selectedTab={section}
                      />
                    )}
                  </SegmenterActions>
                )}
              </EuiPageTemplate.Header>
            </Fragment>
          )}

          <Routes>
            <Route index element={<Navigate to="details" replace={true} />} />
            <Route path="details" element={<SegmentersConfigView segmenter={segmenter} />} />
            <Route path="edit" element={<EditSegmenterView segmenter={segmenter} />} />
          </Routes>
        </Fragment>
      )}
    </Fragment>
  );
};

export default SegmenterDetailsView;
