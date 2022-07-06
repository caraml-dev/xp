import "components/status_badge/StatusBadge.scss";

import React from "react";

import { EuiBadge } from "@elastic/eui";

export const StatusBadge = ({ status }) =>
  !!status ? (
    <EuiBadge
      className="euiBadge--status"
      color={status.color}
      iconType={status.iconType}>
      {status.label}
    </EuiBadge>
  ) : null;
