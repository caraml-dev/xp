import React, { useContext, useEffect, useState } from "react";

import {
  EuiBadge,
  EuiButton,
  EuiFlexItem,
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiSearchBar,
  EuiSpacer,
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { NavigationMenu } from "components/page/NavigationMenu";
import { PageTitle } from "components/page/PageTitle";
import { appConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListSegmentsTable from "./ListSegmentsTable";
import SegmentSearchContext, {
  SegmentSearchContextProvider,
} from "./search/context";

const ListSegmentsComponent = ({ projectId, props }) => {
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const { getFilter, getProcessedFilters, setFilter, isFilterSet } =
    useContext(SegmentSearchContext);

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${projectId}/segments`,
    {
      query: {
        page: page.index + 1,
        page_size: page.size,
        ...getProcessedFilters(),
      },
    },
    { data: [], paging: { total: 0 } }
  );

  useEffect(() => {
    if (isLoaded && !error) {
      setResults({
        items: data.data,
        totalItemCount: data.paging.total,
      });
    }
  }, [data, isLoaded, error]);

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: "../experiments" },
      { text: "Segments" },
    ]);
  }, []);

  const onRowClick = (item) => props.navigate(`./${item.id}/details`);

  return (
    <EuiPage paddingSize="none">
      <EuiPageBody paddingSize="m">
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle
              title="Segments"
              postpend={
                isFilterSet() && <EuiBadge color="primary">Filtered</EuiBadge>
              }
            />
          </EuiPageHeaderSection>
          <EuiPageHeaderSection>
            <NavigationMenu curPage={"segments"} props={props} />
            &emsp;
            <EuiButton size="s" onClick={() => props.navigate("./create")} fill>
              Create Segment
            </EuiButton>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContent>
          <EuiFlexItem grow={6}>
            <EuiSearchBar
              query={getFilter("search") || ""}
              box={{
                placeholder: "Search Segment name",
              }}
              onChange={(text) => {
                setPage({ ...page, index: 0 });
                setFilter("search", text.queryText);
              }}
            />
          </EuiFlexItem>
          <EuiSpacer size="s" />
          <ListSegmentsTable
            isLoaded={isLoaded}
            items={results.items || []}
            page={page}
            error={error}
            onPaginationChange={setPage}
            onRowClick={onRowClick}
            totalItemCount={results.totalItemCount}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

const ListSegmentsView = ({ projectId, ...props }) => (
  <SegmentSearchContextProvider>
    <ListSegmentsComponent projectId={projectId} props={props} />
  </SegmentSearchContextProvider>
);

export default ListSegmentsView;
