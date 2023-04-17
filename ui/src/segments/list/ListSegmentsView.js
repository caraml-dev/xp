import React, { useContext, useEffect, useState } from "react";

import {
  EuiBadge,
  EuiButton,
  EuiFlexItem,
  EuiFlexGroup,
  EuiPanel,
  EuiSearchBar,
  EuiSpacer,
  EuiPageTemplate,
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { useNavigate } from "react-router-dom";

import { NavigationMenu } from "components/page/NavigationMenu";
import { PageTitle } from "components/page/PageTitle";
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";
import NameSearchContext, {
  NameSearchContextProvider,
} from "providers/search/NameSearchContextProvider";

import ListSegmentsTable from "./ListSegmentsTable";

const ListSegmentsComponent = ({ projectId }) => {
  const navigate = useNavigate();
  const {
    appConfig: {
      pagination: { defaultPageSize },
      pageTemplate: { restrictWidth, paddingSize },
    },
  } = useConfig();

  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({
    index: 0,
    size: defaultPageSize,
  });

  const { getFilter, getProcessedFilters, setFilter, isFilterSet } =
    useContext(NameSearchContext);

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

  const onRowClick = (item) => navigate(`./${item.id}/details`);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={
          <PageTitle
            title="Segments"
            postpend={
              isFilterSet() && <EuiBadge color="primary">Filtered</EuiBadge>
            }
          />
        }
        rightSideItems={[
          <EuiButton size="s" onClick={() => navigate("./create")} fill>
            Create Segment
          </EuiButton>,
          <NavigationMenu curPage={"segments"} />,
        ]}
        alignItems={"center"}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <EuiFlexGroup direction="row" gutterSize="s" alignItems="center">
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
          </EuiFlexGroup>
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
        </EuiPanel>
      </EuiPageTemplate.Section>
    </EuiPageTemplate>
  );
};

const ListSegmentsView = ({ projectId }) => (
  <NameSearchContextProvider>
    <ListSegmentsComponent projectId={projectId} />
  </NameSearchContextProvider>
);

export default ListSegmentsView;
