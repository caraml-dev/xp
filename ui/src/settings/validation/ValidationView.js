import { Fragment, useEffect } from "react";

import { EuiSpacer } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ConfigSection } from "components/config_section/ConfigSection";
import { ExternalValidationSection } from "settings/components/config_section/ExternalValidationSection";
import { TreatmentValidationRuleSection } from "settings/components/config_section/TreatmentValidationRuleSection";

const ValidationView = ({ settings }) => {
  const externalValidation = {
    title: "External Validation",
    iconType: "symlink",
    children: (
      <ExternalValidationSection
        settings={settings}></ExternalValidationSection>
    ),
  };
  const treatmentValidationRules = {
    title: "Treatment Validation Rules",
    iconType: "inspect",
    children: (
      <TreatmentValidationRuleSection
        settings={settings}></TreatmentValidationRuleSection>
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
    </Fragment>
  );
};

export default ValidationView;
