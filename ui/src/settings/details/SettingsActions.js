import React, { Fragment, useCallback } from "react";

export const SettingsActions = ({
  onEdit,
  onValidationEdit,
  onCreateSegmenter,
  children,
  selectedTab,
}) => {
  const actions = useCallback(() => {
    return [
      {
        name: "Edit Settings",
        icon: "documentEdit",
        onClick: onEdit,
        hidden: selectedTab !== "details",
      },
      {
        name: "Configure Validation",
        icon: "documentEdit",
        onClick: onValidationEdit,
        hidden: selectedTab !== "validation",
      },
      {
        name: "Create Segmenter",
        icon: "documentEdit",
        onClick: onCreateSegmenter,
        hidden: selectedTab !== "segmenters",
      },
    ];
  }, [onEdit, onValidationEdit, onCreateSegmenter, selectedTab]);

  return <Fragment>{children(actions)}</Fragment>;
};
