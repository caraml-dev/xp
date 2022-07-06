import React, { Fragment, useCallback, useRef } from "react";

import { DeleteSegmenterModal } from "settings/segmenters/details/modals/DeleteSegmenterModal";

export const SegmenterActions = ({ onEdit, onDeleteSuccess, children }) => {
  const deleteSegmentRef = useRef();

  const actions = useCallback(
    (segmenterDetails) => {
      return [
        {
          name: "Edit Segmenter",
          icon: "documentEdit",
          onClick: onEdit,
        },
        {
          name: "Delete Segmenter",
          icon: "trash",
          color: "danger",
          onClick: () => deleteSegmentRef.current(segmenterDetails),
        },
      ];
    },
    [onEdit]
  );

  return (
    <Fragment>
      <DeleteSegmenterModal
        onSuccess={onDeleteSuccess}
        deleteSegmentRef={deleteSegmentRef}
      />
      {children(actions)}
    </Fragment>
  );
};
