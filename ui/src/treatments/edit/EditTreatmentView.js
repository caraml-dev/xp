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
import { EditTreatmentForm } from "treatments/components/form/EditTreatmentForm";

const EditTreatmentView = ({ projectId, treatmentSpec, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Treatments", href: ".." },
      { text: treatmentSpec.name, href: "." },
      { text: "Configuration" },
    ]);
  });

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Edit Treatment" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <FormContextProvider data={Treatment.fromJson(treatmentSpec)}>
            <TreatmentsContextProvider projectId={projectId}>
              <EditTreatmentForm
                projectId={projectId}
                onCancel={() => window.history.back()}
                onSuccess={() => {
                  props.navigate("../", { state: { refresh: true } });
                }}
              />
            </TreatmentsContextProvider>
          </FormContextProvider>
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default EditTreatmentView;
