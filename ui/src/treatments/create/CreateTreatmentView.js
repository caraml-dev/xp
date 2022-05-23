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
import { TreatmentsContextProvider } from "providers/treatment/context";
import { Treatment } from "services/treatment/Treatment";
import { CreateTreatmentForm } from "treatments/components/form/CreateTreatmentForm";

const CreateTreatmentView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Treatments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Treatment" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={new Treatment()}>
            <TreatmentsContextProvider projectId={projectId}>
              <CreateTreatmentForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={(treatmentId) => props.navigate(`../${treatmentId}`)}
              />
            </TreatmentsContextProvider>
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default CreateTreatmentView;
