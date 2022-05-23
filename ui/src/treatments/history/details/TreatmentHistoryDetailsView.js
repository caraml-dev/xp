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
import { useXpApi } from "hooks/useXpApi";
import { GeneralInfoConfigSection } from "treatments/components/configuration/GeneralInfoConfigSection";

const TreatmentHistoryDetailsView = ({ projectId, treatmentId, version }) => {
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
          </Fragment>
        )}
      </EuiPageBody>
    </EuiPage>
  );
};

export default TreatmentHistoryDetailsView;
