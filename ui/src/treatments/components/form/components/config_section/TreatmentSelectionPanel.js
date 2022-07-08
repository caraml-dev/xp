import React, { useContext } from "react";

import {
  EuiFlexGroup,
  EuiLoadingChart,
  EuiSpacer,
  EuiTextAlign,
} from "@elastic/eui";

import { Panel } from "components/panel/Panel";
import TreatmentsContext from "providers/treatment/context";

import { TreatmentConfigPanel } from "./TreatmentConfigPanel";

export const TreatmentSelectionPanel = ({
  projectId,
  treatmentConfig,
  treatmentTemplate,
  onChange,
  errors = [],
}) => {
  const { isLoaded, treatments } = useContext(TreatmentsContext);

  const treatmentSelectionOptions = treatments.map((treatment) => {
    return {
      label: treatment.name,
      id: treatment.id,
    };
  });

  return isLoaded ? (
    <Panel title="Treatment Configuration">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="s">
        <TreatmentConfigPanel
          projectId={projectId}
          treatmentConfig={treatmentConfig}
          treatmentTemplate={treatmentTemplate}
          onChange={onChange}
          treatmentSelectionOptions={treatmentSelectionOptions}
          errors={errors}
        />
      </EuiFlexGroup>
    </Panel>
  ) : (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  );
};
