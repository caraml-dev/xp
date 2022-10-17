import React, { Fragment, useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { Segmenter } from "services/segmenter/Segmenter";
import { CreateSegmenterForm } from "settings/segmenters/components/form/CreateSegmenterForm";

const CreateSegmenterView = () => {
  const { projectId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Segmenters", href: "../segmenters" },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Create Segmenter" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={new Segmenter()}>
          <CreateSegmenterForm
            projectId={projectId}
            onCancel={() => window.history.back()}
            onSuccess={() => navigate(`../`)}
          />
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default CreateSegmenterView;
