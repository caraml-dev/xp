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
import { SegmentsContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { EditSegmentForm } from "segments/components/form/EditSegmentForm";
import { CustomSegment } from "services/segment/CustomSegment";

const EditSegmentView = ({ projectId, segmentSpec, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Segments", href: ".." },
      { text: segmentSpec.name, href: "." },
      { text: "Configuration" },
    ]);
  });

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Edit Segment" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={CustomSegment.fromJson(segmentSpec)}>
            <SegmenterContextProvider projectId={projectId}>
              <SegmentsContextProvider projectId={projectId}>
                <EditSegmentForm
                  projectId={projectId}
                  onCancel={() => window.history.back()}
                  onSuccess={() => {
                    props.navigate("../", { state: { refresh: true } });
                  }}
                />
              </SegmentsContextProvider>
            </SegmenterContextProvider>
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default EditSegmentView;
