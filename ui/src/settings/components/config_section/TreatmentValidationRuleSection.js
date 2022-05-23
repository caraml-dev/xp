import { useRef, useState } from "react";

import { EuiFlexGroup, EuiInMemoryTable } from "@elastic/eui";
import { useDimension, useToggle } from "@gojek/mlp-ui";

import { ConfigPanel } from "components/config_section/ConfigPanel";
import { ConfigSectionFlyout } from "components/config_section/ConfigSectionFlyout";
import { ExpandableTableColumn } from "components/table/ExpandableTableColumn";

export const TreatmentValidationRuleSection = ({ settings }) => {
  const [isFlyoutVisible, toggleFlyout] = useToggle();
  const [flyoutItem, setFlyoutItem] = useState();
  const openFlyout = (item) => () => {
    setFlyoutItem(item);
    toggleFlyout();
  };

  const predicateColumnRef = useRef();
  const { width: contentWidth } = useDimension(predicateColumnRef);
  const columns = [
    {
      field: "name",
      width: "20%",
      name: "Name",
    },
    {
      field: "predicate",
      width: "80%",
      name: "Predicate",
      render: (predicate, item) => {
        return (
          // EuiFlexGroup is used instead of span to measure column length
          <EuiFlexGroup ref={predicateColumnRef} className="eui-textTruncate">
            <ExpandableTableColumn
              text={predicate}
              buttonAction={openFlyout(item)}
              allowedWidth={contentWidth}
            />
          </EuiFlexGroup>
        );
      },
    },
  ];

  return (
    <ConfigPanel>
      <EuiInMemoryTable
        items={settings?.treatment_schema?.rules || []}
        columns={columns}
      />
      {isFlyoutVisible && (
        <ConfigSectionFlyout
          header="Predicate"
          content={flyoutItem.predicate}
          onClose={toggleFlyout}
          size="s"
          contentClass="eui-textBreakWord"
        />
      )}
    </ConfigPanel>
  );
};
