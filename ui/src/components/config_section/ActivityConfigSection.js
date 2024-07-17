import { EuiDescriptionList } from "@elastic/eui";
import { formatDate } from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";

import "./ActivityConfigSection.scss";

export const ActivityConfigSection = ({ spec }) => {
  const items = [
    {
      title: "Created",
      description: formatDate(spec.created_at),
    },
    {
      title: "Last Updated",
      description: formatDate(spec.updated_at),
    },
    {
      title: "Updated By",
      description: spec.updated_by,
    },
  ];

  return (
    <ConfigPanel className="activityConfigPanel">
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        columnWidths={["110px", "auto"]}
      />
    </ConfigPanel>
  );
};
