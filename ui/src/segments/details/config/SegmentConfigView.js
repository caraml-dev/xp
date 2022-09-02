import { Fragment, useEffect } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ActivityConfigSection } from "components/config_section/ActivityConfigSection";
import { ConfigSection } from "components/config_section/ConfigSection";
import { SegmentConfigSection } from "experiments/components/configuration/SegmentConfigSection";
import { SegmenterContextProvider } from "providers/segmenters/context";

export const SegmentConfigView = ({ segment }) => {
  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <ActivityConfigSection spec={segment} />,
  };

  const singleColumnSection = [
    {
      title: "Segment",
      iconType: "package",
      children: (
        <Fragment>
          <SegmenterContextProvider projectId={segment.project_id}>
            <SegmentConfigSection
              experiment={segment}
              projectId={segment.project_id}
            />
          </SegmenterContextProvider>
        </Fragment>
      ),
    },
  ];

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Segments", href: ".." },
      { text: segment.name },
      { text: "Configuration" },
    ]);
  }, [segment]);

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiFlexGroup>
          <EuiFlexItem>
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
