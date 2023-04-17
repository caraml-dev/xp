import React, { Fragment, useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { Segmenter } from "services/segmenter/Segmenter";
import { EditSegmenterForm } from "settings/segmenters/components/form/EditSegmenterForm";

const EditSegmenterView = ({ segmenter }) => {
  const { projectId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Segmenters", href: "../../segmenters" },
      { text: segmenter.name },
      { text: "Configuration" },
    ]);
  }, [segmenter]);

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Segmenter" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={Segmenter.fromJson(segmenter)}>
          <EditSegmenterForm
            projectId={projectId}
            onCancel={() => window.history.back()}
            onSuccess={() =>
              navigate("../", { state: { refresh: true } })
            }
          />
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditSegmenterView;
