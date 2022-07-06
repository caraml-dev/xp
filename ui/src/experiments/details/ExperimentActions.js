import React, { Fragment, useCallback, useRef } from "react";

import { ActivateExperimentModal } from "experiments/details/modals/ActivateExperimentModal";
import { DeactivateExperimentModal } from "experiments/details/modals/DeactivateExperimentModal";

export const ExperimentActions = ({
  onEdit,
  onActivateSuccess,
  onDeactivateSuccess,
  children,
}) => {
  const activateExperimentRef = useRef();
  const deactivateExperimentRef = useRef();

  const actions = useCallback(
    (experiment) => {
      return [
        {
          name: "Edit Experiment",
          icon: "documentEdit",
          onClick: onEdit,
        },
        {
          name: "Activate Experiment",
          icon: "checkInCircleFilled",
          hidden: experiment.status === "active",
          onClick: () => activateExperimentRef.current(experiment),
        },
        {
          name: "Deactivate Experiment",
          icon: "crossInACircleFilled",
          hidden: experiment.status === "inactive",
          onClick: () => deactivateExperimentRef.current(experiment),
        },
      ];
    },
    [onEdit]
  );

  return (
    <Fragment>
      <ActivateExperimentModal
        onSuccess={onActivateSuccess}
        activateExperimentRef={activateExperimentRef}
      />
      <DeactivateExperimentModal
        onSuccess={onDeactivateSuccess}
        deactivateExperimentRef={deactivateExperimentRef}
      />
      {children(actions)}
    </Fragment>
  );
};
