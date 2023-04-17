import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { TreatmentContextProvider } from "providers/treatment/context";
import { Treatment } from "services/treatment/Treatment";
import { CreateTreatmentForm } from "treatments/components/form/CreateTreatmentForm";
import { useConfig } from "config";

const CreateTreatmentView = () => {
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
          <TreatmentContextProvider projectId={projectId}>
            <CreateTreatmentForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={(treatmentId) => navigate(`../${treatmentId}`)}
            />
          </TreatmentContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

export default CreateTreatmentView;
