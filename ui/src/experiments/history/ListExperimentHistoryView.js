import React, { useEffect, useState } from "react";

import { EuiPanel } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ConfigSection } from "components/config_section/ConfigSection";
import { useConfig } from "config";
import ListExperimentHistoryTable from "experiments/history/ListExperimentHistoryTable";
import { useXpApi } from "hooks/useXpApi";

const ListExperimentHistoryView = ({ experiment, ...props }) => {
  const { appConfig } = useConfig();
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${experiment.project_id}/experiments/${experiment.id}/history`,
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
      { text: "Experiments", href: ".." },
      { text: experiment.name, href: "details" },
      { text: "History" },
    ]);
  }, [experiment]);

  return (
    <ConfigSection title="Versions">
      <EuiPanel>
        <ListExperimentHistoryTable
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

export default ListExperimentHistoryView;
