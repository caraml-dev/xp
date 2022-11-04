import React, { Fragment } from "react";

import {
  EuiButtonIcon,
  EuiFieldText,
  EuiFlexItem,
  EuiFormRow,
  EuiText,
} from "@elastic/eui";
import { FormLabelWithToolTip, useToggle } from "@gojek/mlp-ui";

import RouteNamePathFlyout from "./RouteNamePathFlyout";

export const RouteNamePathRow = ({
  routeNamePath,
  routeNamePathPrefix,
  onChange,
  errors,
}) => {
  const [isFlyoutVisible, toggleFlyout] = useToggle();

  return (
    <Fragment>
      <EuiFlexItem>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Route Name Path *"
              content="Specify the path in the treatment configuration where the route name for the final response can be found."
            />
          }
          isInvalid={!!errors}
          error={errors}
          display="row"
        >
          <EuiFieldText
            fullWidth
            placeholder="policy.route_name"
            value={routeNamePath.slice(routeNamePathPrefix.length)}
            onChange={(e) => onChange(routeNamePathPrefix + e.target.value)}
            isInvalid={!!errors}
            name="route-name-path"
            prepend={[
              <EuiButtonIcon
                iconType="questionInCircle"
                onClick={toggleFlyout}
                aria-label="route-name-path-help"
              />,
              <EuiText size={"s"}>{routeNamePathPrefix}</EuiText>,
            ]}
          />
        </EuiFormRow>
      </EuiFlexItem>

      {isFlyoutVisible && (
        <RouteNamePathFlyout
          isFlyoutVisible={isFlyoutVisible}
          onClose={toggleFlyout}
        />
      )}
    </Fragment>
  );
};
