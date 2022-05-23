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
import { CreateSegmentForm } from "segments/components/form/CreateSegmentForm";
import { CustomSegment } from "services/segment/CustomSegment";

const CreateSegmentView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Segments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Segment" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={new CustomSegment()}>
            <SegmenterContextProvider projectId={projectId}>
              <SegmentsContextProvider projectId={projectId}>
                <CreateSegmentForm
                  projectId={projectId}
                  onCancel={() => window.history.back()}
                  onSuccess={(segmentId) => props.navigate(`../${segmentId}`)}
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

export default CreateSegmentView;
