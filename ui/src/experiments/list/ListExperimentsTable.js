import React from "react";

import {
  EuiBasicTable,
  EuiCallOut,
  EuiHealth,
  EuiLink,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
} from "@elastic/eui";

import { useConfig } from "config";
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
  props,
}) => {
  const { appConfig } = useConfig();
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
          <EuiHealth color={experimentStatus.healthColor}>
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
            href={`${props.uri}/${item.id}`}
            target="_blank"></EuiLink>
        );
      },
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

export default ListExperimentsTable;
