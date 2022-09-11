import React from "react";
import { Fragment, useEffect } from "react";

import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiSpacer,
  EuiTextAlign,
  EuiPageTemplate,
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ActivityConfigSection } from "components/config_section/ActivityConfigSection";
import { ConfigSection } from "components/config_section/ConfigSection";
import { PageTitle } from "components/page/PageTitle";
import { SegmentConfigSection } from "experiments/components/configuration/SegmentConfigSection";
import { useXpApi } from "hooks/useXpApi";
import { SegmenterContextProvider } from "providers/segmenters/context";
import { useConfig } from "config";

const SegmentHistoryDetailsView = ({ projectId, segmentId, version }) => {
  const {
    appConfig: {
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [
    {
      data: { data: history },
      isLoaded,
    },
  ] = useXpApi(
    `/projects/${projectId}/segments/${segmentId}/history/${version}`,
    {},
    { data: {} }
  );

  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <ActivityConfigSection spec={history} />,
  };

  const singleColumnSection = [
    {
      title: "Segment",
      iconType: "package",
      children: (
        <Fragment>
          <SegmenterContextProvider projectId={projectId}>
            <SegmentConfigSection experiment={history} projectId={projectId} />
          </SegmenterContextProvider>
        </Fragment>
      ),
    },
  ];

  useEffect(() => {
    isLoaded &&
      replaceBreadcrumbs([
        { text: "Segments", href: "../.." },
        { text: history.name, href: ".." },
        { text: "History", href: "../history" },
        { text: `Version ${history.version}` },
      ]);
  }, [history, isLoaded]);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      {!isLoaded ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : (
        <Fragment>
          <EuiPageTemplate.Header
            bottomBorder={false}
            pageTitle={
              <PageTitle
                title={`${history.name} - Version ${history.version}`}
              />
            }
          />
          <EuiSpacer size="l" />
          <EuiPageTemplate.Section color={"transparent"}>
            <EuiFlexGroup>
              <EuiFlexItem>
                <ConfigSection
                  title={activity.title}
                  iconType={activity.iconType}>
                  {activity.children}
                </ConfigSection>
              </EuiFlexItem>
            </EuiFlexGroup>
            <EuiSpacer size="s" />
            <EuiFlexGroup direction="column">
              {singleColumnSection.map((section, idx) => (
                <EuiFlexItem key={`config-section-${idx}`}>
                  <ConfigSection
                    title={section.title}
                    iconType={section.iconType}>
                    {section.children}
                  </ConfigSection>
                </EuiFlexItem>
              ))}
              <EuiSpacer size="l" />
            </EuiFlexGroup>
          </EuiPageTemplate.Section>
        </Fragment>
      )}
    </EuiPageTemplate>
  );
};

export default SegmentHistoryDetailsView;
