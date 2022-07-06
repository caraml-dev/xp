import "components/config_section/ActivityConfigSection.scss";

import { EuiDescriptionList } from "@elastic/eui";
import { formatDate } from "@elastic/eui/lib/services/format";

import { ConfigPanel } from "components/config_section/ConfigPanel";

export const SegmentersActivitySection = ({ segmenter }) => {
  const items = [
    {
      title: "Created",
      description: formatDate(segmenter.created_at) || "-",
    },
    {
      title: "Last Updated",
      description: formatDate(segmenter.updated_at) || "-",
    },
  ];
  return (
    <ConfigPanel className="activityConfigPanel">
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
      />
    </ConfigPanel>
  );
};
