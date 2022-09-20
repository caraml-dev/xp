import React from "react";

import {
  EuiDescriptionList,
  EuiDescriptionListDescription,
  EuiDescriptionListTitle,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";

export const RouteNamePathConfigGroup = ({ routeNamePath }) => {
  return (
    <EuiTitle size="xs">
      <EuiTextColor>
        <EuiDescriptionList
          textStyle="reverse"
          type="responsiveColumn"
          compressed>
          <EuiDescriptionListTitle>
            Route Name Path
          </EuiDescriptionListTitle>
          <EuiDescriptionListDescription>
            {routeNamePath}
          </EuiDescriptionListDescription>
        </EuiDescriptionList>
      </EuiTextColor>
    </EuiTitle>
  );
};
