import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { PageTitle } from "components/page/PageTitle";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { Settings } from "services/settings/Settings";
import { CreateSettingsForm } from "settings/components/form/CreateSettingsForm";
import { useConfig } from "config";

const CreateSettingsView = ({ projectId, ...props }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings" },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Create Settings" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <SegmenterContextProvider projectId={projectId}>
          <FormContextProvider data={new Settings()}>
            <CreateSettingsForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() => props.navigate(`..`)}
            />
          </FormContextProvider>
        </SegmenterContextProvider>
        <EuiSpacer size="l" />
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default CreateSettingsView;
