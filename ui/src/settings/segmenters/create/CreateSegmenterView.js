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
import { Segmenter } from "services/segmenter/Segmenter";
import { CreateSegmenterForm } from "settings/segmenters/components/form/CreateSegmenterForm";

const CreateSegmenterView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Segmenters", href: "../segmenters" },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Segmenter" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiSpacer size="m" />
        <EuiPageContentBody>
          <FormContextProvider data={new Segmenter()}>
            <CreateSegmenterForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() => props.navigate(`../`)}
            />
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default CreateSegmenterView;
