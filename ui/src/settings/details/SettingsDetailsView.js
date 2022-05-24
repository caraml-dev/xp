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
import { SettingsConfigView } from "settings/details/config/SettingsConfigView";
import { SettingsActions } from "settings/details/SettingsActions";
import EditSettingsView from "settings/edit/EditSettingsView";
import EditValidationView from "settings/edit/EditValidationView";
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
          <>
            {!props["*"].includes("edit") ? (
              <>
                <EuiPageHeader>
                  <EuiPageHeaderSection>
                    <PageTitle icon="managementApp" title="Settings" />
                  </EuiPageHeaderSection>
                </EuiPageHeader>
                <SettingsActions
                  onEdit={() => props.navigate("./edit")}
                  onValidationEdit={() => props.navigate("./validation/edit")}
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
              <EditValidationView
                path="validation/edit"
                settings={data.data}
                isFlyoutVisible={isFlyoutVisible}
                toggleFlyout={toggleFlyout}></EditValidationView>
            </Router>
          </>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default SettingsDetailsView;
