import React, { Fragment } from "react";

import { ConfigSectionTitle } from "./ConfigSectionTitle";

export const ConfigSection = ({ title, iconType, children }) => (
  <Fragment>
    <ConfigSectionTitle title={title} iconType={iconType} />
    {children}
  </Fragment>
);
