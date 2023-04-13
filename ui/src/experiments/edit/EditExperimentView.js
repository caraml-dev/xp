import React, { Fragment, useEffect } from "react";

import { EuiPageTemplate, EuiSpacer } from "@elastic/eui";
import { FormContextProvider, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate } from "react-router-dom";

import { EditExperimentForm } from "experiments/components/form/EditExperimentForm";
import { SegmentContextProvider } from "providers/segment/context";
import { SegmenterContextProvider } from "providers/segmenter/context";
import { SettingsContextProvider } from "providers/settings/context";
import { TreatmentContextProvider } from "providers/treatment/context";
import { Experiment } from "services/experiment/Experiment";
import { PageTitle } from "components/page/PageTitle";

const EditExperimentView = ({ experimentSpec }) => {
  const projectId = experimentSpec.project_id;
  const navigate = useNavigate();
  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: experimentSpec.name, href: "." },
      { text: "Configuration" },
    ]);
  }, [experimentSpec]);

  return (
    <Fragment>
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={<PageTitle title="Edit Experiment" />}
      />
      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <TreatmentContextProvider projectId={projectId}>
          <FormContextProvider data={Experiment.fromJson(experimentSpec)}>
            <SettingsContextProvider projectId={projectId}>
              <SegmenterContextProvider projectId={projectId} status="active">
                <SegmentContextProvider projectId={projectId}>
                  <EditExperimentForm
                    projectId={projectId}
                    onCancel={() => window.history.back()}
                    onSuccess={() =>
                      navigate("../", { state: { refresh: true } })
                    }
                  />
                </SegmentContextProvider>
              </SegmenterContextProvider>
            </SettingsContextProvider>
          </FormContextProvider>
        </TreatmentContextProvider>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default EditExperimentView;
