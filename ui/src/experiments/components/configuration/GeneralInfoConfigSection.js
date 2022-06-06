import { EuiDescriptionList } from "@elastic/eui";
import { formatDate } from "@elastic/eui";
import moment from "moment";

import { ConfigPanel } from "components/config_section/ConfigPanel";
import { useConfig } from "config";

export const GeneralInfoConfigSection = ({ experiment }) => {
  const { appConfig } = useConfig();
  const formatDateValue = (value) =>
    formatDate(
      moment(value, appConfig.datetime.format).utcOffset(
        appConfig.datetime.tzOffsetMinutes
      )
    );
  const items = [
    {
      title: "Description",
      description: experiment?.description || "-",
    },
    {
      title: "Experiment Type",
      description: experiment?.type || "-",
    },
    {
      title: "Experiment Tier",
      description: experiment?.tier,
    },
    {
      title: `Start Time (${appConfig.datetime.tz})`,
      description: formatDateValue(experiment.start_time),
    },
    {
      title: `End Time (${appConfig.datetime.tz})`,
      description: formatDateValue(experiment.end_time),
    },
  ];

  if (experiment?.type === "Switchback") {
    items.splice(2, 0, {
      title: "Switchback Interval",
      description: experiment?.interval || "-",
    });
  }

  return (
    <ConfigPanel>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        titleProps={{ style: { width: "30%" } }}
        descriptionProps={{
          style: { width: "70%", textTransform: "capitalize" },
        }}
      />
    </ConfigPanel>
  );
};
