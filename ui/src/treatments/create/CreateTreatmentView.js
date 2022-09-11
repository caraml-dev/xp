import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { PageTitle } from "components/page/PageTitle";
import { TreatmentsContextProvider } from "providers/treatment/context";
import { Treatment } from "services/treatment/Treatment";
import { CreateTreatmentForm } from "treatments/components/form/CreateTreatmentForm";
import { useConfig } from "config";

const CreateTreatmentView = ({ projectId, ...props }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Treatments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Create Treatment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={new Treatment()}>
          <TreatmentsContextProvider projectId={projectId}>
            <CreateTreatmentForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={(treatmentId) => props.navigate(`../${treatmentId}`)}
            />
          </TreatmentsContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default CreateTreatmentView;
