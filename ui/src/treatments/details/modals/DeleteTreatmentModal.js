import { useEffect, useRef } from "react";

import { ConfirmationModal, addToast } from "@caraml-dev/ui-lib";

import { useModal } from "hooks/useModal";
import { useXpApi } from "hooks/useXpApi";

export const DeleteTreatmentModal = ({ onSuccess, deleteTreatmentRef }) => {
  const closeModalRef = useRef();
  const [treatment = {}, openModal, closeModal] = useModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useXpApi(
    `/projects/${treatment.project_id}/treatments/${treatment.id}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    const _ = require("lodash");
    if (isLoaded && !error && !_.isEmpty(treatment)) {
      addToast({
        id: `submit-success-deactivate-${treatment.name}`,
        title: `Treatment ${treatment.name} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, treatment, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Treatment"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete <b>{treatment.name}</b>.
        </p>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteTreatmentRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
