import React, { Fragment, useCallback, useRef } from "react";

import { DeleteTreatmentModal } from "treatments/details/modals/DeleteTreatmentModal";

export const TreatmentActions = ({ onEdit, onDeleteSuccess, children }) => {
  const deleteTreatmentRef = useRef();

  const actions = useCallback(
    (treatment) => {
      return [
        {
          name: "Edit Treatment",
          icon: "documentEdit",
          onClick: onEdit,
        },
        {
          name: "Delete Treatment",
          icon: "trash",
          color: "danger",
          onClick: () => deleteTreatmentRef.current(treatment),
        },
      ];
    },
    [onEdit]
  );

  return (
    <Fragment>
      <DeleteTreatmentModal
        onSuccess={onDeleteSuccess}
        deleteTreatmentRef={deleteTreatmentRef}
      />
      {children(actions)}
    </Fragment>
  );
};
