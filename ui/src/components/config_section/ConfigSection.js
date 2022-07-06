import React, { Fragment } from "react";

import { ConfigSectionTitle } from "components/config_section/ConfigSectionTitle";

export const ConfigSection = ({ title, iconType, children }) => (
  <Fragment>
    <ConfigSectionTitle title={title} iconType={iconType} />
    {children}
  </Fragment>
);
