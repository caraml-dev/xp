import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { useNavigate, useParams } from "react-router-dom";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";

import { PageTitle } from "components/page/PageTitle";
import { SegmenterContextProvider } from "providers/segmenter/context";
import { Settings } from "services/settings/Settings";
import { CreateSettingsForm } from "settings/components/form/CreateSettingsForm";
import { useConfig } from "config";

const CreateSettingsView = () => {
  const { projectId } = useParams();
  const navigate = useNavigate();

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
              onSuccess={() => navigate(`..`)}
            />
          </FormContextProvider>
        </SegmenterContextProvider>
        <EuiSpacer size="l" />
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default CreateSettingsView;
