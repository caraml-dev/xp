import { Fragment, useEffect } from "react";

import { EuiSpacer, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";

import { ConfigSection } from "components/config_section/ConfigSection";
import { ExternalValidationSection } from "settings/components/config_section/ExternalValidationSection";
import { TreatmentValidationRuleSection } from "settings/components/config_section/TreatmentValidationRuleSection";

const ValidationView = ({ settings }) => {
  const externalValidation = {
    title: "External Validation",
    iconType: "symlink",
    children: (
      <ExternalValidationSection settings={settings} />
    ),
  };
  const treatmentValidationRules = {
    title: "Treatment Validation Rules",
    iconType: "inspect",
    children: (
      <TreatmentValidationRuleSection settings={settings} />
    ),
  };

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings", href: "." },
      { text: "Validation" },
    ]);
  });

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <ConfigSection
          title={externalValidation.title}
          iconType={externalValidation.iconType}>
          {externalValidation.children}
        </ConfigSection>
        <EuiSpacer />
        <ConfigSection
          title={treatmentValidationRules.title}
          iconType={treatmentValidationRules.iconType}>
          {treatmentValidationRules.children}
        </ConfigSection>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default ValidationView;
