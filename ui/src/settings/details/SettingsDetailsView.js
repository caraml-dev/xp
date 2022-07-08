import React, { useEffect } from "react";

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
import { PageNavigation, useToggle } from "@gojek/mlp-ui";
import { Redirect, Router } from "@reach/router";
import classNames from "classnames";

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

import "./SettingsDetailsView.scss";

const SettingsDetailsView = ({ projectId, ...props }) => {
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
    if ((props.location.state || {}).refresh) {
      fetchXPSettings();
    }
  }, [fetchXPSettings, props.location.state]);

  return (
    <EuiPage
      paddingSize="none"
      className={classNames({ pageWithRightSidebar: isFlyoutVisible })}>
      <EuiPageBody paddingSize="m">
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
            {!props["*"].includes("edit") &&
            !props["*"].includes("segmenters/") ? (
              <>
                <EuiPageHeader>
                  <EuiPageHeaderSection>
                    <PageTitle icon="managementApp" title="Settings" />
                  </EuiPageHeaderSection>
                </EuiPageHeader>
                <SettingsActions
                  onEdit={() => props.navigate("./edit")}
                  onValidationEdit={() => props.navigate("./validation/edit")}
                  onCreateSegmenter={() =>
                    props.navigate("./segmenters/create")
                  }
                  selectedTab={props["*"]}>
                  {(getActions) => (
                    <PageNavigation
                      tabs={tabs}
                      actions={getActions()}
                      selectedTab={props["*"]}
                      {...props}
                    />
                  )}
                </SettingsActions>
                <EuiSpacer size="xl" />
              </>
            ) : (
              <></>
            )}
            <Router primary={false}>
              <Redirect from="/" to="details" noThrow />
              <SettingsConfigView path="details" settings={data?.data} />
              <EditSettingsView path="edit" settings={data.data} />
              <ValidationView path="validation" settings={data.data} />
              <CreateSettingsView path="create" />
              <ListSegmentersView path="segmenters" />
              <CreateSegmenterView path="segmenters/create" />
              <SegmenterDetailsView path="segmenters/:segmenterName/*" />
              <EditValidationView
                path="validation/edit"
                settings={data.data}
                isFlyoutVisible={isFlyoutVisible}
                toggleFlyout={toggleFlyout}
              />
            </Router>
          </>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default SettingsDetailsView;
