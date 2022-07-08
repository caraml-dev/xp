import React from "react";

import { EuiHorizontalRule, EuiPanel, EuiTitle } from "@elastic/eui";
import classNames from "classnames";

import "./ConfigPanel.scss";

export const ConfigPanel = ({ title, className, children }) => {
  const classProps = {
    "euiPanel--detailedConfigSection": true,
    [className || ""]: !!className,
  };

  return (
    <EuiPanel className={classNames(classProps)}>
      <>
        {!!title && (
          <>
            <EuiTitle size="xs">
              <span>{title}</span>
            </EuiTitle>
            <EuiHorizontalRule size="full" margin="xs" />
          </>
        )}
        {children}
      </>
    </EuiPanel>
  );
};
