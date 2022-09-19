import React from "react";

import {
  EuiContextMenu,
  EuiPopover,
  EuiButtonEmpty,
} from "@elastic/eui";

export const AffectedExperimentsContextMenu = ({
  item,
  projectId,
  routeToExperimentMappings,
  isButtonPopoverOpen,
  setIsButtonPopoverOpen,
  experimentStatus
}) => {
  console.log(isButtonPopoverOpen);
  let numRoutes = routeToExperimentMappings[item.id] ? Object.keys(routeToExperimentMappings[item.id][experimentStatus]).length : 0;

  const onButtonClick = () => {
    let newIsButtonPopoverOpen = {... isButtonPopoverOpen };
    newIsButtonPopoverOpen[item.id][experimentStatus] = !isButtonPopoverOpen[item.id][experimentStatus];
    setIsButtonPopoverOpen(newIsButtonPopoverOpen);
  };

  const closePopover = () => {
    let newIsButtonPopoverOpen = {... isButtonPopoverOpen };
    newIsButtonPopoverOpen[item.id][experimentStatus] = false;
    setIsButtonPopoverOpen(newIsButtonPopoverOpen);
  };

  const button = (
    <EuiButtonEmpty
      size={"s"}
      iconType={"arrowRight"}
      iconSide={"right"}
      color={"black"}
      onClick={onButtonClick}
      isDisabled={numRoutes === 0}
    >
      {numRoutes}
    </EuiButtonEmpty>
  );

  return isButtonPopoverOpen[item.id] ? (
    <EuiPopover
      button={button}
      isOpen={isButtonPopoverOpen[item.id][experimentStatus]}
      closePopover={closePopover}
      panelPaddingSize={"none"}
      anchorPosition={"rightCenter"}
    >
      <EuiContextMenu
        initialPanelId={0}
        panels={
          [
            {
              id: 0,
              title: `Affected ${experimentStatus[0].toUpperCase() + experimentStatus.slice(1)} Experiments`,
              items: Object.values(routeToExperimentMappings[item.id][experimentStatus]).map(e => (
                {
                  name: e.name,
                  icon: "popout",
                  size: "s",
                  href: `/turing/projects/${projectId}/experiments/${e.id}/details`
                }
              ))
            }
          ]
        }
      />
    </EuiPopover>
  ) : null;
};