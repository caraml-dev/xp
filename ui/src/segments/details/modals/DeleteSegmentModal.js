import { useEffect, useRef } from "react";

import { ConfirmationModal, addToast } from "@gojek/mlp-ui";

import { useModal } from "hooks/useModal";
import { useXpApi } from "hooks/useXpApi";

export const DeleteSegmentModal = ({ onSuccess, deleteSegmentRef }) => {
  const closeModalRef = useRef();
  const [segment = {}, openModal, closeModal] = useModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useXpApi(
    `/projects/${segment.project_id}/segments/${segment.id}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    const _ = require("lodash");
    if (isLoaded && !error && !_.isEmpty(segment)) {
      addToast({
        id: `submit-success-deactivate-${segment.name}`,
        title: `Segment ${segment.name} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, segment, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Segment"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete <b>{segment.name}</b>.
        </p>
      }
      confirmButtonText="Delete"
      confirmButtonColor="danger">
      {(onSubmit) =>
        (deleteSegmentRef.current = openModal(onSubmit)) &&
        (closeModalRef.current = onSubmit) && <span />
      }
    </ConfirmationModal>
  );
};
