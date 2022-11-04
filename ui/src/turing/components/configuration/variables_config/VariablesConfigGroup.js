import React from "react";

import { EuiBasicTable, EuiText, EuiTextColor, EuiTitle } from "@elastic/eui";

import { mapProtocolLabel } from "turing/components/utils/helper";

export const VariablesConfigGroup = ({ variables, protocol }) => {
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
      render: (value) => (
        <EuiText size="s">{mapProtocolLabel(protocol, value)}</EuiText>
      ),
    },
  ];

  return <EuiBasicTable items={variables} columns={columns} itemId="name" />;
};
