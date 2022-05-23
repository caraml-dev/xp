import React from "react";
import { Fragment, useEffect } from "react";

import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiPage,
  EuiPageBody,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ActivityConfigSection } from "components/config_section/ActivityConfigSection";
import { ConfigSection } from "components/config_section/ConfigSection";
import { PageTitle } from "components/page/PageTitle";
import { GeneralInfoConfigSection } from "experiments/components/configuration/GeneralInfoConfigSection";
import { SegmentConfigSection } from "experiments/components/configuration/SegmentConfigSection";
import { TreatmentConfigSection } from "experiments/components/configuration/TreatmentConfigSection";
import { useXpApi } from "hooks/useXpApi";
import { SegmenterContextProvider } from "providers/segmenters/context";

const ExperimentHistoryDetailsView = ({ projectId, experimentId, version }) => {
  const [
    {
      data: { data: history },
      isLoaded,
    },
  ] = useXpApi(
    `/projects/${projectId}/experiments/${experimentId}/history/${version}`,
    {},
    { data: {} }
  );

  const generalInfo = {
    title: "General Info",
    iconType: "apmTrace",
    children: <GeneralInfoConfigSection experiment={history} />,
  };

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
    {
      title: "Treatments",
      iconType: "beaker",
      children: <TreatmentConfigSection experiment={history} />,
    },
  ];

  useEffect(() => {
    isLoaded &&
      replaceBreadcrumbs([
        { text: "Experiments", href: "../.." },
        { text: history.name, href: ".." },
        { text: "History", href: "../history" },
        { text: `Version ${history.version}` },
      ]);
  }, [history, isLoaded]);

  return (
    <EuiPage>
      <EuiPageBody>
        {!isLoaded ? (
          <EuiTextAlign textAlign="center">
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        ) : (
          <Fragment>
            <EuiPageHeader>
              <EuiPageHeaderSection>
                <PageTitle
                  title={`${history.name} - Version ${history.version}`}
                />
              </EuiPageHeaderSection>
            </EuiPageHeader>
            <Fragment>
              <EuiFlexGroup direction="row">
                <EuiFlexItem grow={2}>
                  <ConfigSection
                    title={generalInfo.title}
                    iconType={generalInfo.iconType}>
                    {generalInfo.children}
                  </ConfigSection>
                </EuiFlexItem>
                <EuiFlexItem grow={1}>
                  <ConfigSection
                    title={activity.title}
                    iconType={activity.iconType}>
                    {activity.children}
                  </ConfigSection>
                </EuiFlexItem>
              </EuiFlexGroup>
              <EuiSpacer size="l" />
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
            </Fragment>
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default ExperimentHistoryDetailsView;
