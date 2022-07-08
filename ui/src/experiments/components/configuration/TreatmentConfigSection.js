import { React, useRef, useState } from "react";

import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiInMemoryTable,
  EuiPanel,
  EuiText,
} from "@elastic/eui";
import { useDimension, useToggle } from "@gojek/mlp-ui";

import { ConfigPanel } from "components/config_section/ConfigPanel";
import { ConfigSectionFlyout } from "components/config_section/ConfigSectionFlyout";
import { ExpandableTableColumn } from "components/table/ExpandableTableColumn";
import { formatJsonString } from "utils/helpers";

export const TreatmentConfigSection = ({ experiment }) => {
  const [isFlyoutVisible, toggleFlyout] = useToggle();

  const [flyoutItem, setFlyoutItem] = useState();
  const openFlyout = (configuration) => () => {
    setFlyoutItem(formatJsonString(configuration));
    toggleFlyout();
  };

  const treatmentColumnRef = useRef();
  const { width: contentWidth } = useDimension(treatmentColumnRef);

  const columns = [
    {
      field: "name",
      width: "20%",
      name: "Name",
    },
    {
      field: "traffic",
      width: "20%",
      name: "Traffic",
      render: (traffic) => {
        return (
          <EuiText
            size="s"
            style={{ fontWeight: "bold" }}
            className="eui-textTruncate ">
            {traffic}
          </EuiText>
        );
      },
    },
    {
      field: "configuration",
      name: "Configuration",
      width: "50%",
      render: (configuration, item) => {
        return (
          <EuiFlexGroup ref={treatmentColumnRef} className="eui-textTruncate">
            <ExpandableTableColumn
              text={JSON.stringify(configuration)}
              buttonAction={openFlyout(item)}
              allowedWidth={contentWidth}
            />
          </EuiFlexGroup>
        );
      },
    },
  ];

  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem>
        {experiment.treatments ? (
          <ConfigPanel>
            <EuiInMemoryTable items={experiment.treatments} columns={columns} />
            {isFlyoutVisible && (
              <ConfigSectionFlyout
                header="Configuration"
                content={flyoutItem}
                onClose={toggleFlyout}
                size="m"
                contentClass="eui-textBreakWord"
              />
            )}
          </ConfigPanel>
        ) : (
          <EuiPanel>
            <EuiText size="s">Not Configured</EuiText>
          </EuiPanel>
        )}
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
