import React, { useState } from "react";

import { EuiButtonIcon, EuiContextMenu, EuiPopover } from "@elastic/eui";
import { useNavigate } from "react-router-dom";

export const NavigationMenu = ({ curPage }) => {
  const navigate = useNavigate();
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
      onClick: () => navigate("../experiments/settings"),
    },
    {
      name: "Experiments",
      icon: "apmTrace",
      onClick: () => navigate("../experiments"),
    },
    {
      name: "Treatments",
      icon: "beaker",
      onClick: () => navigate("../experiments/treatments"),
    },
    {
      name: "Segments",
      icon: "package",
      onClick: () => navigate("../experiments/segments"),
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
