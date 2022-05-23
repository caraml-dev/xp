import React, { useEffect } from "react";

import {
  EuiPage,
  EuiPageBody,
  EuiPageContentBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { PageTitle } from "components/page/PageTitle";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { Settings } from "services/settings/Settings";
import { CreateSettingsForm } from "settings/components/form/CreateSettingsForm";

const CreateSettingsView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings" },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Settings" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <SegmenterContextProvider>
            <FormContextProvider data={new Settings()}>
              <CreateSettingsForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={() => props.navigate(`..`)}
              />
            </FormContextProvider>
          </SegmenterContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default CreateSettingsView;
