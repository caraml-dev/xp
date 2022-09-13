import React, { Fragment, useEffect } from "react";

import {
  EuiPageTemplate,
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
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Treatment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
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
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditTreatmentView;
