import React from "react";

import {
  EuiBasicTable,
  EuiCallOut,
  EuiLoadingChart,
  EuiTextAlign,
} from "@elastic/eui";

const ListTreatmentsTable = ({
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
      field: "id",
      name: "ID",
      width: "5%",
    },
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
      field: "updated_at",
      name: "Updated At",
      dataType: "date",
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
      itemId="id"
      pagination={pagination}
      onChange={onTableChange}
    />
  );
};

export default ListTreatmentsTable;
