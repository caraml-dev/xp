import React from "react";

import {
  EuiDescriptionList,
  EuiTitle,
} from "@elastic/eui";

export const RouteNamePathConfigGroup = ({ routeNamePath }) => {
  return (
    <EuiTitle size="xs">
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={[
          {
            title: "Route Name Path",
            description: routeNamePath,
          },
        ]}
        columnWidths={[1, 7/3]}
      />
    </EuiTitle>
  );
};
