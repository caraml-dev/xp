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
import { CreateExperimentForm } from "experiments/components/form/CreateExperimentForm";
import { SegmentsContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { SettingsContextProvider } from "providers/settings/context";
import { TreatmentsContextProvider } from "providers/treatment/context";
import { Experiment } from "services/experiment/Experiment";

const CreateExperimentView = ({ projectId, ...props }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "." },
      { text: "Create" },
    ]);
  }, [projectId]);

  return (
    <EuiPage>
      <EuiPageBody>
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle title="Create Experiment" />
          </EuiPageHeaderSection>
        </EuiPageHeader>
        <EuiPageContentBody>
          <TreatmentsContextProvider projectId={projectId}>
            <FormContextProvider data={new Experiment()}>
              <SettingsContextProvider projectId={projectId}>
                <SegmenterContextProvider projectId={projectId}>
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
          <EuiSpacer size="l" />
        </EuiPageContentBody>
      </EuiPageBody>
    </EuiPage>
  );
};

export default CreateExperimentView;
