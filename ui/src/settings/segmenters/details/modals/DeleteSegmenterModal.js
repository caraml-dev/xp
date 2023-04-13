import { useEffect, useRef } from "react";

import { ConfirmationModal, addToast } from "@caraml-dev/ui-lib";

import { useModal } from "hooks/useModal";
import { useXpApi } from "hooks/useXpApi";

export const DeleteSegmenterModal = ({ onSuccess, deleteSegmentRef }) => {
  const closeModalRef = useRef();
  const [segmenterDetails = {}, openModal, closeModal] =
    useModal(closeModalRef);

  const [{ isLoading, isLoaded, error }, submitForm] = useXpApi(
    `/projects/${segmenterDetails.projectId}/segmenters/${segmenterDetails.name}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
    },
    {},
    false
  );

  useEffect(() => {
    const _ = require("lodash");
    if (isLoaded && !error && !_.isEmpty(segmenterDetails)) {
      addToast({
        id: `submit-success-delete-${segmenterDetails.name}`,
        title: `Segmenter ${segmenterDetails.name} has been deleted!`,
        color: "success",
        iconType: "check",
      });
      onSuccess();
      closeModal();
    }
  }, [isLoaded, error, segmenterDetails, onSuccess, closeModal]);

  return (
    <ConfirmationModal
      title="Delete Segmenter"
      onConfirm={submitForm}
      isLoading={isLoading}
      content={
        <p>
          You are about to delete <b>{segmenterDetails.name}</b>.
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
