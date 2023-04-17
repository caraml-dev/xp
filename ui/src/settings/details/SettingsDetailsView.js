import React, { useEffect } from "react";

import {
  EuiCallOut,
  EuiLoadingChart,
  EuiPageTemplate,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { PageNavigation, useToggle } from "@caraml-dev/ui-lib";
import { Navigate, Routes, Route, useLocation, useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { useXpApi } from "hooks/useXpApi";
import CreateSettingsView from "settings/create/CreateSettingsView";
import { SettingsConfigView } from "settings/details/config/SettingsConfigView";
import { SettingsActions } from "settings/details/SettingsActions";
import EditSettingsView from "settings/edit/EditSettingsView";
import EditValidationView from "settings/edit/EditValidationView";
import CreateSegmenterView from "settings/segmenters/create/CreateSegmenterView";
import SegmenterDetailsView from "settings/segmenters/details/SegmenterDetailsView";
import { ListSegmentersView } from "settings/segmenters/list/ListSegmentersView";
import ValidationView from "settings/validation/ValidationView";

import { useConfig } from "config";

const SettingsDetailsView = () => {
  const { projectId, "*": section } = useParams();
  const location = useLocation();
  const navigate = useNavigate();
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [isFlyoutVisible, toggleFlyout] = useToggle();
  const [{ data, isLoaded, error }, fetchXPSettings] = useXpApi(
    `/projects/${projectId}/settings`
  );

  const tabs = [
    {
      id: "details",
      name: "General",
    },
    {
      id: "validation",
      name: "Validation",
    },
    {
      id: "segmenters",
      name: "Segmenters",
    },
  ];

  useEffect(() => {
    if ((location.state || {}).refresh) {
      fetchXPSettings();
    }
  }, [fetchXPSettings, location.state]);

  return (
    <EuiPageTemplate
      restrictWidth={restrictWidth}
      paddingSize={paddingSize}
    >
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
        <>
          {!section.includes("edit") &&
            !section.includes("segmenters/") && (
              <>
                <EuiPageTemplate.Header
                  bottomBorder={false}
                  pageTitle={<PageTitle icon="managementApp" title="Settings" />}
                >
                  <SettingsActions
                    onEdit={() => navigate("./edit")}
                    onValidationEdit={() => navigate("./validation/edit")}
                    onCreateSegmenter={() =>
                      navigate("./segmenters/create")
                    }
                    selectedTab={section}>
                    {(getActions) => (
                      <PageNavigation
                        tabs={tabs}
                        actions={getActions()}
                        selectedTab={section}
                      />
                    )}
                  </SettingsActions>
                </EuiPageTemplate.Header>
              </>
            )}

          <Routes>
            {/* DETAILS */}
            <Route index element={<Navigate to="details" replace={true} />} />
            <Route path="details" element={<SettingsConfigView settings={data?.data} />} />
            {/* CREATE */}
            <Route path="create" element={<CreateSettingsView />} />
            {/* EDIT */}
            <Route path="edit" element={<EditSettingsView settings={data.data} />} />
            {/* VALIDATION */}
            <Route path="validation">
              <Route index element={<ValidationView settings={data.data} />} />
              <Route path="edit" element={<EditValidationView settings={data.data} isFlyoutVisible={isFlyoutVisible} toggleFlyout={toggleFlyout} />} />
            </Route>
            {/* SEGMENTER */}
            <Route path="segmenters">
              <Route index element={<ListSegmentersView />} />
              <Route path="create" element={<CreateSegmenterView />} />
              <Route path=":segmenterName/*" element={<SegmenterDetailsView />} />
            </Route>
          </Routes>
        </>
      )}
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default SettingsDetailsView;
