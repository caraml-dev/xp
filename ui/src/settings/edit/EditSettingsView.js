import React, { Fragment, useEffect } from "react";

import { EuiPageTemplate, EuiSpacer } from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { SegmenterContextProvider } from "providers/segmenters/context";
import { Settings } from "services/settings/Settings";
import { EditSettingsForm } from "settings/components/form/EditSettingsForm";
import { PageTitle } from "components/page/PageTitle";

const EditSettingsView = ({ projectId, settings, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings", href: "." },
      { text: "Configuration" },
    ]);
  });

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Settings" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={Settings.fromJson(settings)}>
          <SegmenterContextProvider projectId={projectId}>
            <EditSettingsForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() => {
                props.navigate("../", { state: { refresh: true } });
              }}
            />
          </SegmenterContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditSettingsView;
