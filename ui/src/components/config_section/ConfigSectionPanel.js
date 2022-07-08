import React from "react";

import { ConfigMultiSectionPanel } from "./ConfigMultiSectionPanel";

export const ConfigSectionPanel = (props) => (
  <ConfigMultiSectionPanel
    items={[{ title: props.title, children: props.children }]}
    className={props.className}
  />
);
