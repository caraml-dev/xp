import React, { Fragment, useContext } from "react";

import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiLoadingChart,
  EuiTextAlign,
} from "@elastic/eui";
import {
  FormContext,
  FormValidationContext,
  get,
  useOnChangeHandler,
} from "@caraml-dev/ui-lib";

import { TreatmentsConfigPanel } from "experiments/components/form/components/treatments_config/TreatmentsConfigPanel";

export const TreatmentsStep = ({ projectId }) => {
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
        <EuiFlexGroup direction="column" gutterSize="m">
          <EuiFlexItem grow={true}>
            <TreatmentsConfigPanel
              projectId={projectId}
              treatments={data.treatments}
              onChange={onChange}
              errors={get(errors, "treatments")}
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      )}
    </Fragment>
  );
};
