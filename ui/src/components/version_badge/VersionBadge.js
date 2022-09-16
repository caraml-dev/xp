import React from "react";

import { EuiBadge } from "@elastic/eui";

export const VersionBadge = ({ version }) => (
  <EuiBadge
    color={"hollow"}>
    {`v${version}`}
  </EuiBadge>
);