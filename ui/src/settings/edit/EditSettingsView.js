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
import { EditSettingsForm } from "settings/components/form/EditSettingsForm";

const EditSettingsView = ({ projectId, settings, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings", href: "." },
      { text: "Configuration" },
    ]);
  });

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Edit Settings" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={Settings.fromJson(settings)}>
            <SegmenterContextProvider>
              <EditSettingsForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={() => {
                  props.navigate("../", { state: { refresh: true } });
                }}
              />
            </SegmenterContextProvider>
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default EditSettingsView;
