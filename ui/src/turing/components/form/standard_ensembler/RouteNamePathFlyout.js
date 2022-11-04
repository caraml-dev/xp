import React from "react";

import {
  EuiCode,
  EuiCodeBlock,
  EuiFlexItem,
  EuiFlyout,
  EuiFlyoutBody,
  EuiFlyoutHeader,
  EuiText,
  EuiTitle,
} from "@elastic/eui";

import { useConfig } from "config";

const RouteNamePathFlyout = ({ onClose }) => {
  const {
    appConfig: { routeNamePathPrefix },
  } = useConfig();

  return (
    <EuiFlyout ownFocus onClose={onClose} size={"s"} paddingSize="m">
      <EuiFlyoutHeader hasBorder>
        <EuiFlexItem>
          <EuiTitle size="s">
            <h4>Route Name Path Prefix</h4>
          </EuiTitle>
        </EuiFlexItem>
      </EuiFlyoutHeader>
      <EuiFlyoutBody>
        <EuiText>
          <EuiFlexItem>
            <EuiText>
              <p>
                The prefix in the grayed-out area specifies the path prefix that
                gets appended to a user-defined treatment configuration.
              </p>

              <p>
                This path prefix reflects the nesting of the treatment
                configuration within the final response payload that the Turing
                Router finally receives.
              </p>

              <p>
                In Turing Experiments' case, if the user-defined treatment
                configuration is:
              </p>
              <EuiCodeBlock language="json" fontSize="m" paddingSize="m">
                {`{
    "route_name": "control",
    ...
}`}
              </EuiCodeBlock>

              <p>
                the client response that gets sent back to the Turing Router is
                actually as follows:
              </p>
              <EuiCodeBlock language="json" fontSize="m" paddingSize="m">
                {`{
    "treatment": {
        "configuration": {
            "route_name": "control"
            ...
        }
        ....
    }
    ...
}`}
              </EuiCodeBlock>

              <p>
                Hence, the path prefix is automatically specified as
                <EuiCode language="json">{routeNamePathPrefix}</EuiCode>{" "}
                (including the final period).
              </p>
            </EuiText>
          </EuiFlexItem>
        </EuiText>
      </EuiFlyoutBody>
    </EuiFlyout>
  );
};

export default RouteNamePathFlyout;
