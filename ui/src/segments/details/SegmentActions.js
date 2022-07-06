import React, { Fragment, useCallback, useRef } from "react";

import { DeleteSegmentModal } from "segments/details/modals/DeleteSegmentModal";

export const SegmentActions = ({ onEdit, onDeleteSuccess, children }) => {
  const deleteSegmentRef = useRef();

  const actions = useCallback(
    (segment) => {
      return [
        {
          name: "Edit Segment",
          icon: "documentEdit",
          onClick: onEdit,
        },
        {
          name: "Delete Segment",
          icon: "trash",
          color: "danger",
          onClick: () => deleteSegmentRef.current(segment),
        },
      ];
    },
    [onEdit]
  );

  return (
    <Fragment>
      <DeleteSegmentModal
        onSuccess={onDeleteSuccess}
        deleteSegmentRef={deleteSegmentRef}
      />
      {children(actions)}
    </Fragment>
  );
};
