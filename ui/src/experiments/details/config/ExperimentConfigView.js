import React, { Fragment, useEffect } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ActivityConfigSection } from "components/config_section/ActivityConfigSection";
import { ConfigSection } from "components/config_section/ConfigSection";
import { GeneralInfoConfigSection } from "experiments/components/configuration/GeneralInfoConfigSection";
import { SegmentConfigSection } from "experiments/components/configuration/SegmentConfigSection";
import { TreatmentConfigSection } from "experiments/components/configuration/TreatmentConfigSection";
import { SegmenterContextProvider } from "providers/segmenters/context";

export const ExperimentConfigView = ({ experiment }) => {
  const generalInfo = {
    title: "General Info",
    iconType: "apmTrace",
    children: <GeneralInfoConfigSection experiment={experiment} />,
  };

  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <ActivityConfigSection spec={experiment} />,
  };

  const singleColumnSection = [
    {
      title: "Segment",
      iconType: "package",
      children: (
        <Fragment>
          <SegmenterContextProvider projectId={experiment.project_id}>
            <SegmentConfigSection
              experiment={experiment}
              projectId={experiment.project_id}
            />
          </SegmenterContextProvider>
        </Fragment>
      ),
    },
    {
      title: "Treatments",
      iconType: "beaker",
      children: <TreatmentConfigSection experiment={experiment} />,
    },
  ];

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: experiment.name },
      { text: "Configuration" },
    ]);
  }, [experiment]);

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <ConfigSection
              title={generalInfo.title}
              iconType={generalInfo.iconType}>
              {generalInfo.children}
            </ConfigSection>
          </EuiFlexItem>
          <EuiFlexItem grow={1}>
            <ConfigSection title={activity.title} iconType={activity.iconType}>
              {activity.children}
            </ConfigSection>
          </EuiFlexItem>
        </EuiFlexGroup>
        <EuiSpacer size="l" />
        <EuiFlexGroup direction="column">
          {singleColumnSection.map((section, idx) => (
            <EuiFlexItem key={`config-section-${idx}`}>
              <ConfigSection title={section.title} iconType={section.iconType}>
                {section.children}
              </ConfigSection>
            </EuiFlexItem>
          ))}
          <EuiSpacer size="l" />
        </EuiFlexGroup>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};
