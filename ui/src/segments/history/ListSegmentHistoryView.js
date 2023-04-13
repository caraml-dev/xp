import React, { Fragment, useEffect, useState } from "react";

import { EuiPanel, EuiSpacer, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate } from "react-router-dom";

import { ConfigSection } from "components/config_section/ConfigSection";
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListSegmentHistoryTable from "./ListSegmentHistoryTable";

const ListSegmentHistoryView = ({ segment }) => {
  const navigate = useNavigate();
  const { appConfig } = useConfig();
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${segment.project_id}/segments/${segment.id}/history`,
    {
      query: {
        page: page.index + 1,
        page_size: page.size,
      },
    },
    { data: [], paging: { total: 0 } }
  );

  const onRowClick = (item) => navigate(`./${item.version}`);

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Segments", href: ".." },
      { text: segment.name, href: "details" },
      { text: "History" },
    ]);
  }, [segment]);

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <ConfigSection title="Versions">
          <EuiPanel>
            <ListSegmentHistoryTable
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
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export default ListSegmentHistoryView;
