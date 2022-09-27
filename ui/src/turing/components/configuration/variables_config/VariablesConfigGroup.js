import React from "react";

import { EuiBasicTable, EuiTextColor, EuiTitle } from "@elastic/eui";

export const VariablesConfigGroup = ({ variables }) => {
  const columns = [
    {
      field: "name",
      name: "Variable",
      width: "35%",
      render: (value) => (
        <EuiTitle size="xxs">
          <EuiTextColor color="secondary">{value}</EuiTextColor>
        </EuiTitle>
      ),
    },
    {
      field: "field",
      name: "Field",
      width: "35%",
      render: (value) => (
        <EuiTitle size="xxs">
          <EuiTextColor>{value || "-"}</EuiTextColor>
        </EuiTitle>
      ),
    },
    {
      field: "field_source",
      name: "Source",
      width: "30%",
    },
  ];

  return <EuiBasicTable items={variables} columns={columns} itemId="name" />;
};
