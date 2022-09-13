import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";

import { PageTitle } from "components/page/PageTitle";
import { CreateExperimentForm } from "experiments/components/form/CreateExperimentForm";
import { SegmentsContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { SettingsContextProvider } from "providers/settings/context";
import { TreatmentsContextProvider } from "providers/treatment/context";
import { Experiment } from "services/experiment/Experiment";
import { useConfig } from "config";

const CreateExperimentView = ({ projectId, ...props }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Create Experiment" />}
      />

      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <TreatmentsContextProvider projectId={projectId}>
          <FormContextProvider data={new Experiment()}>
            <SettingsContextProvider projectId={projectId}>
              <SegmenterContextProvider projectId={projectId} status="active">
                <SegmentsContextProvider projectId={projectId}>
                  <CreateExperimentForm
                    projectId={projectId}
                    onCancel={() => window.history.back()}
                    onSuccess={(experimentId) =>
                      props.navigate(`../${experimentId}`)
                    }
                  />
                </SegmentsContextProvider>
              </SegmenterContextProvider>
            </SettingsContextProvider>
          </FormContextProvider>
        </TreatmentsContextProvider>
      </EuiPageTemplate.Section>

      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default CreateExperimentView;
