import React from "react";

import {
  EuiBasicTable,
  EuiCallOut,
  EuiLoadingChart,
  EuiTextAlign,
} from "@elastic/eui";

const ListTreatmentHistoryTable = ({
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
    showPerPageOptions: false,
    totalItemCount,
  };

  const columns = [
    {
      field: "version",
      name: "Version",
      width: "5%",
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

export default ListTreatmentHistoryTable;
