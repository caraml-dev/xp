import React, { useEffect, useState } from "react";

import { EuiPanel } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ConfigSection } from "components/config_section/ConfigSection";
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListTreatmentHistoryTable from "./ListTreatmentHistoryTable";

const ListTreatmentHistoryView = ({ treatment, ...props }) => {
  const { appConfig } = useConfig();
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${treatment.project_id}/treatments/${treatment.id}/history`,
    {
      query: {
        page: page.index + 1,
        page_size: page.size,
      },
    },
    { data: [], paging: { total: 0 } }
  );

  const onRowClick = (item) => props.navigate(`./${item.version}`);

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Treatments", href: ".." },
      { text: treatment.name, href: "details" },
      { text: "History" },
    ]);
  }, [treatment]);

  return (
    <ConfigSection title="Versions">
      <EuiPanel>
        <ListTreatmentHistoryTable
          items={data.data}
          isLoaded={isLoaded}
          error={error}
          page={page}
          totalItemCount={data.paging.total}
          onPaginationChange={setPage}
          onRowClick={onRowClick}
        />
      </EuiPanel>
    </ConfigSection>
  );
};

export default ListTreatmentHistoryView;
