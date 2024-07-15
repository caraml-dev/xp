import { EuiDescriptionList } from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";

export const ExternalValidationSection = ({ settings }) => {
  const items = [
    {
      title: "Url",
      description: settings?.validation_url || "-",
    },
  ];

  return (
    <ConfigPanel>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        columnWidths={[1, 4]}
      />
    </ConfigPanel>
  );
};
