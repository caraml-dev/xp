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
import { useXpApi } from "hooks/useXpApi";
import { GeneralInfoConfigSection } from "treatments/components/configuration/GeneralInfoConfigSection";
import { useConfig } from "config";

const TreatmentHistoryDetailsView = ({ projectId, treatmentId, version }) => {
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
    `/projects/${projectId}/treatments/${treatmentId}/history/${version}`,
    {},
    { data: {} }
  );

  const generalInfo = {
    title: "Configuration",
    iconType: "apmTrace",
    children: <GeneralInfoConfigSection treatment={history} />,
  };

  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <ActivityConfigSection spec={history} />,
  };

  useEffect(() => {
    isLoaded &&
      replaceBreadcrumbs([
        { text: "Treatments", href: "../.." },
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
              <EuiFlexGroup>
                <EuiFlexItem>
                  <ConfigSection
                    title={generalInfo.title}
                    iconType={generalInfo.iconType}>
                    {generalInfo.children}
                  </ConfigSection>
                </EuiFlexItem>
              </EuiFlexGroup>
            </EuiPageTemplate.Section>
          </Fragment>
        )}
    </EuiPageTemplate>
  );
};

export default TreatmentHistoryDetailsView;
