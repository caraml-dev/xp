import React, { useEffect } from "react";

import {
  EuiPageTemplate,
  EuiSpacer,
} from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useNavigate, useParams } from "react-router-dom";

import { PageTitle } from "components/page/PageTitle";
import { CreateExperimentForm } from "experiments/components/form/CreateExperimentForm";
import { SegmentContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenter/context";
import { SettingsContextProvider } from "providers/settings/context";
import { TreatmentContextProvider } from "providers/treatment/context";
import { Experiment } from "services/experiment/Experiment";
import { useConfig } from "config";

const CreateExperimentView = () => {
  const { projectId } = useParams();
  const navigate = useNavigate();
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
        <TreatmentContextProvider projectId={projectId}>
          <FormContextProvider data={new Experiment()}>
            <SettingsContextProvider projectId={projectId}>
              <SegmenterContextProvider projectId={projectId} status="active">
                <SegmentContextProvider projectId={projectId}>
                  <CreateExperimentForm
                    projectId={projectId}
                    onCancel={() => window.history.back()}
                    onSuccess={(experimentId) =>
                      navigate(`../${experimentId}`)
                    }
                  />
                </SegmentContextProvider>
              </SegmenterContextProvider>
            </SettingsContextProvider>
          </FormContextProvider>
        </TreatmentContextProvider>
      </EuiPageTemplate.Section>

      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

export default CreateExperimentView;
