import React from "react";

import { EuiHealth } from "@elastic/eui";

import { BasicTable } from "components/table/BasicTable";
import { getSegmenterStatus } from "services/segmenter/SegmenterStatus";
import { formatDateCell } from "utils/helpers";

export const ListSegmentersTable = ({ items, isLoaded, error, onRowClick }) => {
  const columns = [
    {
      field: "name",
      name: "Name",
      mobileOptions: {
        enlarge: true,
        fullWidth: true,
      },
      width: "20%",
    },
    {
      field: "type",
      name: "Type",
      width: "5%",
    },
    {
      field: "scope",
      name: "Scope",
      width: "5%",
    },
    {
      field: "status",
      name: "Status",
      width: "5%",
      render: (_, item) => {
        var segmenterStatus = getSegmenterStatus(item);
        return (
          <EuiHealth color={segmenterStatus.color}>
            {segmenterStatus.label}
          </EuiHealth>
        );
      },
    },
    {
      field: "required",
      name: "Required",
      width: "5%",
    },
    {
      field: "created_at",
      name: "Created At",
      dataType: "date",
      width: "10%",
      align: "center",
      render: formatDateCell,
    },
    {
      field: "updated_at",
      name: "Updated At",
      dataType: "date",
      width: "10%",
      align: "center",
      render: formatDateCell,
    },
  ];

  return (
    <BasicTable
      items={items}
      isLoaded={isLoaded}
      error={error}
      tableColumns={columns}
      onRowClick={onRowClick}
    />
  );
};
