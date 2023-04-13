import React, { Fragment, useEffect } from "react";

import {
  EuiButton,
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { Settings } from "services/settings/Settings";
import { EditValidationForm } from "settings/components/form/EditValidationForm";
import PlaygroundFlyout from "settings/components/playground_flyout/PlaygroundFlyout";

const EditValidationView = ({
  settings,
  isFlyoutVisible,
  toggleFlyout,
}) => {
  const { projectId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Settings", href: ".." },
      { text: "Validation", href: "." },
      { text: "Edit" },
    ]);
  });

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Validation" />}
        rightSideItems={[
          <EuiButton size="s" onClick={toggleFlyout}>
            Playground
          </EuiButton>
        ]}
        alignItems={"center"}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={Settings.fromJson(settings)}>
          <EditValidationForm
            projectId={projectId}
            onCancel={() => window.history.back()}
            onSuccess={() => {
              navigate("../", { state: { refresh: true } });
            }}
          />

          {isFlyoutVisible && (
            <PlaygroundFlyout
              isFlyoutVisible={isFlyoutVisible}
              onClose={toggleFlyout}
            />
          )}
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditValidationView;
