import React, { Fragment, useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { TreatmentContextProvider } from "providers/treatment/context";
import { Treatment } from "services/treatment/Treatment";
import { EditTreatmentForm } from "treatments/components/form/EditTreatmentForm";

const EditTreatmentView = ({ treatmentSpec }) => {
  const { projectId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Treatments", href: ".." },
      { text: treatmentSpec.name, href: "." },
      { text: "Configuration" },
    ]);
  });

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Treatment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <FormContextProvider data={Treatment.fromJson(treatmentSpec)}>
          <TreatmentContextProvider projectId={projectId}>
            <EditTreatmentForm
              projectId={projectId}
              onCancel={() => window.history.back()}
              onSuccess={() => {
                navigate("../", { state: { refresh: true } });
              }}
            />
          </TreatmentContextProvider>
        </FormContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditTreatmentView;
