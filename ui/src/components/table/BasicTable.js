import {
  EuiBasicTable,
  EuiCallOut,
  EuiLoadingChart,
  EuiTextAlign,
} from "@elastic/eui";

export const BasicTable = ({
  items,
  isLoaded,
  error,
  page,
  totalItemCount,
  tableColumns,
  onRowClick,
  onPaginationChange,
}) => {
  const cellProps = (item) =>
    !!onRowClick
      ? {
          style: { cursor: "pointer" },
          onClick: () => onRowClick(item),
        }
      : undefined;

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
  ) : !!onPaginationChange && !!totalItemCount && !!page ? (
    <EuiBasicTable
      items={items}
      columns={tableColumns}
      cellProps={cellProps}
      itemId="id"
      pagination={{
        pageIndex: page.index,
        pageSize: page.size,
        showPerPageOptions: false,
        totalItemCount,
      }}
      onChange={({ page = {} }) => onPaginationChange(page)}
    />
  ) : (
    <EuiBasicTable
      items={items}
      columns={tableColumns}
      cellProps={cellProps}
      itemId="id"
    />
  );
};
