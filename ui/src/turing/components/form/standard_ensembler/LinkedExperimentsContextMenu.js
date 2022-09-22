import React from "react";

import {
  EuiContextMenu,
  EuiPopover,
  EuiButtonEmpty,
  EuiTextColor,
  EuiIcon
} from "@elastic/eui";

export const LinkedExperimentsContextMenu = ({
  item,
  projectId,
  routeToExperimentMappings,
  isButtonPopoverOpen,
  setIsButtonPopoverOpen,
  experimentStatus
}) => {
  let numRoutes = routeToExperimentMappings[item.id] ? Object.keys(routeToExperimentMappings[item.id][experimentStatus]).length : 0;

  const onButtonClick = () => {
    let newIsButtonPopoverOpen = { ...isButtonPopoverOpen };
    newIsButtonPopoverOpen[item.id][experimentStatus] = !isButtonPopoverOpen[item.id][experimentStatus];
    setIsButtonPopoverOpen(newIsButtonPopoverOpen);
  };

  const closePopover = () => {
    let newIsButtonPopoverOpen = { ...isButtonPopoverOpen };
    newIsButtonPopoverOpen[item.id][experimentStatus] = false;
    setIsButtonPopoverOpen(newIsButtonPopoverOpen);
  };

  const button = (
    <EuiButtonEmpty
      size={"s"}
      iconType={"arrowRight"}
      iconSide={"right"}
      color={"primary"}
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
              title: `Linked ${experimentStatus[0].toUpperCase() + experimentStatus.slice(1)} Experiments`,
              items: Object.values(routeToExperimentMappings[item.id][experimentStatus]).map(e => (
                {
                  name: (
                    <EuiTextColor>
                      <a href={`/turing/projects/${projectId}/experiments/${e.id}/details`} target={"_blank"}>
                        {e.name}
                      </a>
                    </EuiTextColor>
                  ),
                  icon: <EuiIcon type={"popout"} size={"m"} color={"primary"} />,
                  size: "s",
                  toolTipContent: "Open experiment details page",
                  toolTipPosition: "right",
                }
              ))
            }
          ]
        }
      />
    </EuiPopover>
  ) : null;
};