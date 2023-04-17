import React from "react";

import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiSpacer } from "@elastic/eui";
import { get, useOnChangeHandler } from "@caraml-dev/ui-lib";

import { Panel } from "components/panel/Panel";
import { newConstraint } from "services/segmenter/Segmenter";
import { ConstraintCard } from "settings/segmenters/components/form/components/ConstraintCard";

export const ConstraintsPanel = ({
  constraints,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const onAddConstraint = () => {
    onChange("constraints")([...constraints, newConstraint()]);
  };

  const onDeleteConstraint = (idx) => () => {
    constraints.splice(idx, 1);
    onChange("constraints")(constraints);
  };

  return (
    <Panel title={"Constraints"}>
      <EuiFlexGroup direction="column" gutterSize="s">
        {constraints.map((constraint, idx) => (
          <EuiFlexItem key={`constraint-${idx}`}>
            <ConstraintCard
              constraint={constraint}
              onChangeHandler={onChange(`constraints.${idx}`)}
              onDelete={onDeleteConstraint(idx)}
              errors={get(errors, `${idx}`)}
            />
            <EuiSpacer size="s" />
          </EuiFlexItem>
        ))}
        <EuiFlexItem>
          <EuiButton fullWidth onClick={onAddConstraint}>
            + Add Constraint
          </EuiButton>
        </EuiFlexItem>
      </EuiFlexGroup>
    </Panel>
  );
};
