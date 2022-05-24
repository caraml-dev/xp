import React, { Fragment, useCallback } from "react";

export const SettingsActions = ({
  onEdit,
  onValidationEdit,
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
    ];
  }, [onEdit, onValidationEdit, selectedTab]);

  return <Fragment>{children(actions)}</Fragment>;
};
