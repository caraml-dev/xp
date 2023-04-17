import React, { Fragment, useContext } from "react";

import { EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { TreatmentValidationRulesConfigPanel } from "settings/components/form/components/treatment_validation_section/TreatmentValidationRulesConfigPanel";

export const TreatmentValidationRulesStep = () => {
  const { data, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const { errors } = useContext(FormValidationContext);

  return (
    <Fragment>
      {!data ? (
        <EuiTextAlign textAlign="center">
          <EuiLoadingChart size="xl" mono />
        </EuiTextAlign>
      ) : (
        <TreatmentValidationRulesConfigPanel
          settings={data}
          onChange={onChange}
          errors={errors}
        />
      )}
    </Fragment>
  );
};
