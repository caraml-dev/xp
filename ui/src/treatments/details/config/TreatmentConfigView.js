import { Fragment, useEffect } from "react";

import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

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
    </Fragment>
  );
};
