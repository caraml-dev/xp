import { EuiDescriptionList } from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";

export const SegmentersGeneralSection = ({ segmenter }) => {
  const items = [
    {
      title: "Description",
      description: segmenter?.description || "-",
    },
    {
      title: "Type",
      description: segmenter.type,
    },
    {
      title: "Required",
      description: segmenter.required.toString(),
    },
    {
      title: "Multi Valued",
      description: segmenter.multi_valued.toString(),
    },
    {
      title: "Scope",
      description: segmenter.scope,
    },
  ];
  return (
    <ConfigPanel>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        columnWidths={[1, 7/3]}
      />
    </ConfigPanel>
  );
};
