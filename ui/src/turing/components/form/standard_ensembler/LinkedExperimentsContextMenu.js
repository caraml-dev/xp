import React from "react";

import {
  EuiButtonEmpty,
  EuiContextMenu,
  EuiIcon,
  EuiLink,
  EuiPopover,
} from "@elastic/eui";
import { useToggle } from "@gojek/mlp-ui";

export const LinkedExperimentsContextMenu = ({
  projectId,
  linkedExperiments,
  experimentStatus,
}) => {
  const [isPopoverOpen, togglePopover] = useToggle();

  let numExperiments = Object.keys(linkedExperiments[experimentStatus]).length;

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

  return (
    <EuiPopover
      button={button}
      isOpen={isPopoverOpen}
      closePopover={togglePopover}
      panelPaddingSize={"none"}
      anchorPosition={"rightCenter"}
    >
      <EuiContextMenu
        initialPanelId={0}
        panels={[
          {
            id: 0,
            title: `Linked ${
              experimentStatus[0].toUpperCase() + experimentStatus.slice(1)
            } Experiments`,
            items: Object.values(linkedExperiments[experimentStatus]).map(
              (e) => ({
                name: (
                  <EuiLink
                    href={`/turing/projects/${projectId}/experiments/${e.id}/details`}
                    external={false}
                    target={"_blank"}
                  >
                    {e.name}
                  </EuiLink>
                ),
                icon: <EuiIcon type={"popout"} size={"m"} color={"primary"} />,
                size: "s",
                toolTipContent: "Open experiment details page",
                toolTipPosition: "right",
              })
            ),
          },
        ]}
      />
    </EuiPopover>
  );
};
