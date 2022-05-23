import React, { useCallback, useEffect } from "react";

import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";

import { Panel } from "components/panel/Panel";
import { TreatmentCard } from "experiments/components/form/components/treatments_config/TreatmentCard";
import { makeNewTreatment } from "utils/helpers";

export const TreatmentsConfigPanel = ({
  projectId,
  treatments,
  onChange,
  errors = [],
}) => {
  const retrieveTreatmentIdxByUuid = (uuid) => {
    return treatments.findIndex((treatment) => treatment.uuid === uuid);
  };

  //Create a new Treatment Card whenever list is empty
  useEffect(() => {
    if (!treatments.length) {
      onChange("treatments")([...treatments, makeNewTreatment()]);
    }
  }, [onChange, treatments]);

  // Initialize name and valueType as empty strings to avoid Uncontrolled input error
  const onAddTreatment = useCallback(() => {
    onChange("treatments")([...treatments, { ...makeNewTreatment() }]);
  }, [onChange, treatments]);

  const onDeleteTreatment = (idx) => () => {
    treatments.splice(idx, 1);
    onChange("treatments")([...treatments]);
  };

  const onUpdateTreatment = (uuid, key, value) => {
    treatments[retrieveTreatmentIdxByUuid(uuid)][key] = value;
    onChange("treatments")([...treatments]);
  };

  return (
    <Panel title="Treatment Configuration">
      <EuiSpacer size="s" />
      <EuiFlexGroup direction="column" gutterSize="s">
        {treatments.map((treatment, idx) => (
          <EuiFlexItem key={idx}>
            <TreatmentCard
              projectId={projectId}
              treatment={treatment}
              onUpdateTreatment={onUpdateTreatment}
              onDelete={
                treatments.length > 1 ? onDeleteTreatment(idx) : undefined
              }
              errors={errors[idx]}
            />
            <EuiSpacer size="s" />
          </EuiFlexItem>
        ))}
      </EuiFlexGroup>
      <EuiSpacer />
      <EuiFlexItem>
        <EuiButton onClick={onAddTreatment}>+ Add Treatment</EuiButton>
      </EuiFlexItem>
    </Panel>
  );
};
