import React, { useEffect } from "react";

import {
  EuiPageTemplate,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { EditExperimentForm } from "experiments/components/form/EditExperimentForm";
import { SegmentsContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { SettingsContextProvider } from "providers/settings/context";
import { TreatmentsContextProvider } from "providers/treatment/context";
import { Experiment } from "services/experiment/Experiment";

const EditExperimentView = ({ projectId, experimentSpec, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: experimentSpec.name, href: "." },
      { text: "Configuration" },
    ]);
  }, [experimentSpec]);

  return (
    <EuiPageTemplate.Section color={"transparent"}>
      <TreatmentsContextProvider projectId={projectId}>
        <FormContextProvider data={Experiment.fromJson(experimentSpec)}>
          <SettingsContextProvider projectId={projectId}>
            <SegmenterContextProvider projectId={projectId} status="active">
              <SegmentsContextProvider projectId={projectId}>
                <EditExperimentForm
                  projectId={projectId}
                  onCancel={() => window.history.back()}
                  onSuccess={() =>
                    props.navigate("../", { state: { refresh: true } })
                  }
                />
              </SegmentsContextProvider>
            </SegmenterContextProvider>
          </SettingsContextProvider>
        </FormContextProvider>
      </TreatmentsContextProvider>
    </EuiPageTemplate.Section>
  );
};

export default EditExperimentView;
