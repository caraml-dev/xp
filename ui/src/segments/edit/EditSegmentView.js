import { Fragment, useEffect } from "react";

import {
  EuiPageTemplate,
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
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Segment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={CustomSegment.fromJson(segmentSpec)}>
          <SegmenterContextProvider projectId={projectId} status="active">
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
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditSegmentView;
