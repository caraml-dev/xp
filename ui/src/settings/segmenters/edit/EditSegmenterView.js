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
import { EditSegmenterForm } from "settings/segmenters/components/form/EditSegmenterForm";

const EditSegmenterView = ({ projectId, segmenter, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Segmenters", href: "../../segmenters" },
      { text: segmenter.name },
      { text: "Configuration" },
    ]);
  }, [segmenter]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Edit Segmenter" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiSpacer size="m" />
        <EuiPageContentBody>
          <FormContextProvider data={Segmenter.fromJson(segmenter)}>
            <EditSegmenterForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() =>
                props.navigate("../", { state: { refresh: true } })
              }
            />
          </FormContextProvider>
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default EditSegmenterView;
