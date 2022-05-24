import { useEffect, useRef } from "react";

import { ConfirmationModal, addToast } from "@gojek/mlp-ui";

import { useModal } from "hooks/useModal";
import { useXpApi } from "hooks/useXpApi";

export const DeactivateExperimentModal = ({
  onSuccess,
  deactivateExperimentRef,
}) => {
  const closeModalRef = useRef();
  const [experiment = {}, openModal, closeModal] = useModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useXpApi(
    `/projects/${experiment.project_id}/experiments/${experiment.id}/disable`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    if (isLoaded && !error) {
      addToast({
        id: `submit-success-deactivate-${experiment.name}`,
        title: `Experiment ${experiment.name} has been deactivated!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, experiment, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Deactivate Experiment"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to deactivate <b>{experiment.name}</b>.
        </p>
      }
      confirmButtonText="Deactivate"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deactivateExperimentRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
