import React from "react";

import {
  EuiBasicTable,
  EuiCallOut,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
} from "@elastic/eui";

import { appConfig } from "config";
import { formatDateCell } from "utils/helpers";

const ListExperimentHistoryTable = ({
  items,
  isLoaded,
  error,
  page,
  totalItemCount,
  onPaginationChange,
  onRowClick,
}) => {
  const pagination = {
    pageIndex: page.index,
    pageSize: page.size,
    totalItemCount,
    hidePerPageOptions: true,
  };

  const columns = [
    {
      field: "version",
      name: "Version",
      width: "5%",
    },
    {
      field: "status",
      name: "Status",
      width: "5%",
      render: (value) => (
        <EuiText size="s" style={{ textTransform: "capitalize" }}>
          {value}
        </EuiText>
      ),
    },
    {
      field: "tier",
      name: "Tier",
      width: "5%",
      render: (value) => (
        <EuiText size="s" style={{ textTransform: "capitalize" }}>
          {value}
        </EuiText>
      ),
    },
    {
      field: "start_time",
      name: `Start Time (${appConfig.datetime.tz})`,
      width: "8%",
      render: formatDateCell,
    },
    {
      field: "end_time",
      name: `End Time (${appConfig.datetime.tz})`,
      width: "8%",
      render: formatDateCell,
    },
    {
      field: "created_at",
      name: "Created At",
      dataType: "date",
      width: "8%",
    },
    {
      field: "updated_at",
      name: "Updated At",
      dataType: "date",
      width: "8%",
    },
    {
      field: "updated_by",
      name: "Updated By",
      width: "10%",
    },
  ];

  const cellProps = (item) =>
    onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: () => onRowClick(item),
        }
      : undefined;

  const onTableChange = ({ page = {} }) => onPaginationChange(page);

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <EuiBasicTable
      items={items}
      columns={columns}
      cellProps={cellProps}
      itemId="version"
      pagination={pagination}
      onChange={onTableChange}
    />
  );
};

export default ListExperimentHistoryTable;
