import { EuiDescriptionList, EuiFieldPassword, EuiPanel } from "@elastic/eui";
import { formatDate } from "@elastic/eui";

import "./Password.scss";

export const GeneralInfoSection = ({ settings }) => {
  const items = [
    {
      title: "Name",
      description: settings.username,
    },
    {
      title: "Passkey",
      description: (
        <EuiFieldPassword
          value={settings.passkey}
          compressed={true}
          readOnly={true}
          type="dual"
        />
      ),
    },
    {
      title: "Created At",
      description: formatDate(settings.created_at),
    },
    {
      title: "Updated At",
      description: formatDate(settings.updated_at),
    },
  ];

  return (
    <EuiPanel>
      <EuiDescriptionList
        compressed
        textStyle="reverse"
        type="responsiveColumn"
        listItems={items}
        columnWidths={[1, 7/3]}
      />
    </EuiPanel>
  );
};
