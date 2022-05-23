import React, { useState } from "react";

import { EuiButtonIcon, EuiContextMenu, EuiPopover } from "@elastic/eui";

export const NavigationMenu = ({ curPage, props }) => {
  const [isPopoverOpen, setPopover] = useState(false);

  const onButtonClick = () => {
    setPopover(!isPopoverOpen);
  };

  const closePopover = () => {
    setPopover(false);
  };

  const allRows = [
    {
      name: "Settings",
      icon: "gear",
      onClick: () => props.navigate("../experiments/settings"),
    },
    {
      name: "Experiments",
      icon: "apmTrace",
      onClick: () => props.navigate("../experiments"),
    },
    {
      name: "Treatments",
      icon: "beaker",
      onClick: () => props.navigate("../experiments/treatments"),
    },
    {
      name: "Segments",
      icon: "package",
      onClick: () => props.navigate("../experiments/segments"),
    },
  ];
  const contextRows = allRows.filter((e) => e.name.toLowerCase() !== curPage);

  const panels = [
    {
      id: 0,
      title: "Navigation Menu",
      items: contextRows,
    },
  ];

  const button = (
    <EuiButtonIcon
      onClick={onButtonClick}
      display="base"
      iconType="arrowDown"
      size="s"
      iconSize="l"
      aria-label="navigation"
    />
  );

  return (
    <EuiPopover
      id="nav-context-menu"
      button={button}
      isOpen={isPopoverOpen}
      closePopover={closePopover}
      panelPaddingSize="none"
      anchorPosition="downLeft">
      <EuiContextMenu initialPanelId={0} panels={panels} />
    </EuiPopover>
  );
};
