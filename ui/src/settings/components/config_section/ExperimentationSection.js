import {
  EuiDescriptionList,
  EuiFlexGroup,
  EuiFlexItem,
  EuiInMemoryTable,
  EuiText,
} from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";

const SegmentersSection = ({ segmenters }) => {
  const columns = [
    {
      field: "name",
      width: "40%",
      name: "Name",
    },
    {
      field: "variables",
      width: "60%",
      name: "Experiment Variable(s)",
    },
  ];
  const items = segmenters.names.map((name) => ({
    name,
    variables: segmenters.variables[name],
  }));

  return (
    <ConfigPanel title={"Segmenters"}>
      {items.length > 0 ? (
        <EuiInMemoryTable items={items} columns={columns}></EuiInMemoryTable>
      ) : (
        <EuiText>Not Configured</EuiText>
      )}
    </ConfigPanel>
  );
};

const RandomizationSection = ({ settings }) => {
  const items = [
    {
      title: "Randomization Key",
      description: settings.randomization_key,
    },
  ];
  return (
    <ConfigPanel title={"Randomization"}>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        columnWidths={[1, 1]}
      />
    </ConfigPanel>
  );
};

export const ExperimentationSection = ({ settings }) => (
  <EuiFlexGroup direction="row">
    <EuiFlexItem grow={2}>
      <SegmentersSection segmenters={settings.segmenters} />
    </EuiFlexItem>
    <EuiFlexItem grow={1}>
      <RandomizationSection settings={settings} />
    </EuiFlexItem>
  </EuiFlexGroup>
);
