import { Fragment, useEffect } from "react";

import { EuiFlexGroup, EuiFlexItem, EuiSpacer, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";

import { ActivityConfigSection } from "components/config_section/ActivityConfigSection";
import { ConfigSection } from "components/config_section/ConfigSection";
import { GeneralInfoConfigSection } from "treatments/components/configuration/GeneralInfoConfigSection";

export const TreatmentConfigView = ({ treatment }) => {
  const generalInfo = {
    title: "Configuration",
    iconType: "apmTrace",
    children: <GeneralInfoConfigSection treatment={treatment} />,
  };

  const activity = {
    title: "Activity",
    iconType: "indexEdit",
    children: <ActivityConfigSection spec={treatment} />,
  };

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../.." },
      { text: "Treatments", href: ".." },
      { text: treatment.name },
      { text: "Configuration" },
    ]);
  }, [treatment]);

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
  );
};
