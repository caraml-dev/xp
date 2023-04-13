import { useEffect, useRef } from "react";

import { ConfirmationModal, addToast } from "@caraml-dev/ui-lib";

import { useModal } from "hooks/useModal";
import { useXpApi } from "hooks/useXpApi";

export const ActivateExperimentModal = ({
  onSuccess,
  activateExperimentRef,
}) => {
  const closeModalRef = useRef();
  const [experiment = {}, openModal, closeModal] = useModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useXpApi(
    `/projects/${experiment.project_id}/experiments/${experiment.id}/enable`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    const _ = require("lodash");
    if (isLoaded && !error) {
      addToast({
        id: `submit-success-activate-${experiment.name}`,
        title: `Experiment ${experiment.name} has been activated!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
    if (!_.isEmpty(experiment) && error) {
      closeModal();
    }
  }, [isLoaded, error, experiment, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Activate Experiment"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to activate <b>{experiment.name}</b>.
        </p>
      }
      confirmButtonText="Activate"
      confirmButtonColor="primary">
      {(onSubmit) =>
        (activateExperimentRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
