import React from "react";

import {
  EuiContextMenu,
  EuiPopover,
  EuiButtonEmpty,
  EuiLink,
  EuiIcon
} from "@elastic/eui";
import { useToggle } from "@gojek/mlp-ui";

export const LinkedExperimentsContextMenu = ({
  projectId,
  linkedExperiments,
  experimentStatus
}) => {
  const [isPopoverOpen, togglePopover] = useToggle();

  let numExperiments = linkedExperiments ? Object.keys(linkedExperiments[experimentStatus]).length : 0;

  const button = (
    <EuiButtonEmpty
      size={"s"}
      iconType={"arrowRight"}
      iconSide={"right"}
      color={"primary"}
      onClick={togglePopover}
      isDisabled={numExperiments === 0}
    >
      {numExperiments}
    </EuiButtonEmpty>
  );

  return linkedExperiments ? (
    <EuiPopover
      button={button}
      isOpen={isPopoverOpen}
      closePopover={togglePopover}
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
              items: Object.values(linkedExperiments[experimentStatus]).map(e => (
                {
                  name: (
                    <EuiLink href={`/turing/projects/${projectId}/experiments/${e.id}/details`} external={false} target={"_blank"}>
                      {e.name}
                    </EuiLink>
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