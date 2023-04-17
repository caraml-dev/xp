import React, { useContext } from "react";

import { EuiRadioGroup } from "@elastic/eui";
import { FormContext } from "@caraml-dev/ui-lib";

import { getValidationOptions } from "settings/components/playground_flyout/typeOptions";

export const PlaygroundRadioGroup = ({
  selectedValidationType,
  setSelectedValidationType,
}) => {
  const { data: settings } = useContext(FormContext);

  return (
    <EuiRadioGroup
      options={getValidationOptions(settings).map((e) => ({
        id: e.id,
        label: e.label,
        disabled: e.disabled,
      }))}
      idSelected={selectedValidationType}
      onChange={(optionId) => {
        setSelectedValidationType(optionId);
      }}
      name="validation radio group"
      legend={{
        children: <span>Validation Type</span>,
      }}
    />
  );
};
