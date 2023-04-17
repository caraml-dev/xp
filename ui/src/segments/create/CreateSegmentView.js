import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { SegmentContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenter/context";
import { CreateSegmentForm } from "segments/components/form/CreateSegmentForm";
import { CustomSegment } from "services/segment/CustomSegment";
import { useConfig } from "config";

const CreateSegmentView = () => {
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
      { text: "Segments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Create Segment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={new CustomSegment()}>
          <SegmenterContextProvider projectId={projectId} status="active">
            <SegmentContextProvider projectId={projectId}>
              <CreateSegmentForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={(segmentId) => navigate(`../${segmentId}`)}
              />
            </SegmentContextProvider>
          </SegmenterContextProvider>
        </FormContextProvider>
        <EuiSpacer size="l" />
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default CreateSegmentView;
