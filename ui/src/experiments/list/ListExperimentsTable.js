import React from "react";

import { EuiHealth, EuiLink, EuiText } from "@elastic/eui";
import { useLocation } from "react-router-dom";

import { useConfig } from "config";
import { BasicTable } from "components/table/BasicTable";
import { getExperimentStatus } from "services/experiment/ExperimentStatus";
import { formatDateCell } from "utils/helpers";

const ListExperimentsTable = ({
  items,
  isLoaded,
  error,
  page,
  totalItemCount,
  onPaginationChange,
  onRowClick,
}) => {
  const location = useLocation();
  const { appConfig } = useConfig();
  const tableColumns = [
    {
      field: "id",
      name: "ID",
      width: "5%",
    },
    {
      field: "type",
      name: "Type",
      width: "7%",
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
      field: "status",
      name: "Status",
      width: "8%",
      render: (_, item) => {
        var experimentStatus = getExperimentStatus(item);
        return (
          <EuiHealth color={experimentStatus.color}>
            {experimentStatus.label}
          </EuiHealth>
        );
      },
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
      width: "10%",
      render: formatDateCell,
    },
    {
      field: "end_time",
      name: `End Time (${appConfig.datetime.tz})`,
      width: "10%",
      render: formatDateCell,
    },
    {
      field: "updated_at",
      name: "Updated At",
      dataType: "date",
      width: "10%",
    },
    {
      name: "Actions",
      align: "right",
      mobileOptions: {
        header: true,
        fullWidth: false,
      },
      width: "5%",
      render: (item) => {
        return (
          <EuiLink
            onClick={(e) => {
              e.stopPropagation();
            }}
            href={`${location.uri}/${item.id}`}
            target="_blank"
          />
        );
      },
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

export default ListExperimentsTable;
