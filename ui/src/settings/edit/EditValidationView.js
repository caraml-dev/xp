import React, { useEffect } from "react";

import {
  EuiButton,
  EuiPage,
  EuiPageBody,
  EuiPageContentBody,
  EuiPageHeader,
  EuiPageHeaderSection,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";
import classNames from "classnames";

import { PageTitle } from "components/page/PageTitle";
import { Settings } from "services/settings/Settings";
import { EditValidationForm } from "settings/components/form/EditValidationForm";
import PlaygroundFlyout from "settings/components/playground_flyout/PlaygroundFlyout";

const EditValidationView = ({
  projectId,
  settings,
  isFlyoutVisible,
  toggleFlyout,
  ...props
}) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Settings", href: ".." },
      { text: "Validation", href: "." },
      { text: "Edit" },
    ]);
  });

  return (
    <FormContextProvider data={Settings.fromJson(settings)}>
      <EuiPage
        paddingSize="none"
        className={classNames({ pageWithRightSidebar: isFlyoutVisible })}>
        <EuiPageBody paddingSize="m">
          <EuiPageHeader>
            <EuiPageHeaderSection>
              <PageTitle title="Edit Validation" />
            </EuiPageHeaderSection>
            <EuiPageHeaderSection>
              <EuiButton size="s" onClick={toggleFlyout}>
                Playground
              </EuiButton>
            </EuiPageHeaderSection>
          </EuiPageHeader>

          <EuiPageContentBody>
            <EditValidationForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() => {
                props.navigate("../", { state: { refresh: true } });
              }}
            />
          </EuiPageContentBody>
        </EuiPageBody>

        {isFlyoutVisible && (
          <PlaygroundFlyout
            isFlyoutVisible={isFlyoutVisible}
            onClose={toggleFlyout}
          />
        )}
      </EuiPage>
    </FormContextProvider>
  );
};

export default EditValidationView;
