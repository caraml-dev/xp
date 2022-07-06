import React from "react";

import { ConfigMultiSectionPanel } from "components/config_section/ConfigMultiSectionPanel";

export const ConfigSectionPanel = (props) => (
  <ConfigMultiSectionPanel
    items={[{ title: props.title, children: props.children }]}
    className={props.className}
  />
);
