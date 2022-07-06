import React from "react";

import { BasicTable } from "components/table/BasicTable";

const ListTreatmentsTable = ({
  items,
  isLoaded,
  error,
  page,
  totalItemCount,
  onPaginationChange,
  onRowClick,
}) => {
  const tableColumns = [
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

  return (
    <BasicTable
      items={items}
      isLoaded={isLoaded}
      error={error}
      page={page}
      totalItemCount={totalItemCount}
      tableColumns={tableColumns}
      onPaginationChange={onPaginationChange}
      onRowClick={onRowClick}
    />
  );
};

export default ListTreatmentsTable;
