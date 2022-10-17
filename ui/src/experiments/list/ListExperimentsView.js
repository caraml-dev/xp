import React, { useContext, useEffect, useState } from "react";

import {
  EuiBadge,
  EuiButton,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSearchBar,
  EuiSpacer,
  EuiPageTemplate
} from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { useNavigate } from "react-router-dom";

import { NavigationMenu } from "components/page/NavigationMenu";
import { PageTitle } from "components/page/PageTitle";
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListExperimentsTable from "./ListExperimentsTable";
import ExperimentSearchContext, {
  ExperimentSearchContextProvider,
} from "./search/context";
import SearchExperimentsPanel from "./search/SearchExperimentsPanel";

const ListExperimentsComponent = ({ projectId }) => {
  const navigate = useNavigate();
  const {
    appConfig: {
      pagination: { defaultPageSize },
      pageTemplate: { restrictWidth, paddingSize },
      experimentsTableColumns: experimentsTableFields,
    },
  } = useConfig();

  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({
    index: 0,
    size: defaultPageSize,
  });
  // Search related states
  const [isSearchPanelVisible, setIsSearchPanelVisible] = useState(false);
  const { getFilter, getProcessedFilters, setFilter, isFilterSet } = useContext(
    ExperimentSearchContext
  );

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      query: {
        page: page.index + 1,
        page_size: page.size,
        fields: experimentsTableFields,
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
    replaceBreadcrumbs([{ text: "Experiments" }]);
  }, []);

  const onRowClick = (item) => navigate(`./${item.id}/details`);

  return (
    <EuiPageTemplate restrictWidth={restrictWidth} paddingSize={paddingSize}>
      <EuiSpacer size="l" />
      {isSearchPanelVisible && (
        <SearchExperimentsPanel
          onChange={() => setPage({ ...page, index: 0 })}
          onClose={setIsSearchPanelVisible}
          projectId={projectId}
        />
      )}

      <EuiPageTemplate.Header
        bottomBorder={false}
        pageTitle={
          <PageTitle
            title="Experiments"
            postpend={
              isFilterSet() && <EuiBadge color="primary">Filtered</EuiBadge>
            }
          />
        }
        rightSideItems={[
          <EuiButton size="s" onClick={() => navigate("./create")} fill>
            Create Experiment
          </EuiButton>,
          <NavigationMenu curPage={"experiments"} />,
        ]}
        alignItems={"center"}
      />

      <EuiSpacer size="l" />
      <EuiPageTemplate.Section color={"transparent"}>
        <EuiPanel>
          <EuiFlexGroup direction="row" gutterSize="s" alignItems="center">
            <EuiFlexItem grow={false}>
              <EuiButton
                size="s"
                onClick={() => setIsSearchPanelVisible(!isSearchPanelVisible)}>
                {"Search Options"}
              </EuiButton>
            </EuiFlexItem>
            <EuiFlexItem>
              <EuiSearchBar
                query={getFilter("search") || ""}
                box={{
                  placeholder: "Search Experiment name or description",
                }}
                onChange={(text) => {
                  setFilter("search", text.queryText);
                }}
              />
            </EuiFlexItem>
          </EuiFlexGroup>
          <EuiSpacer size="s" />
          <ListExperimentsTable
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
      <EuiSpacer size="l" />
    </EuiPageTemplate>
  );
};

const ListExperimentsView = ({ projectId }) => (
  <ExperimentSearchContextProvider>
    <ListExperimentsComponent projectId={projectId} />
  </ExperimentSearchContextProvider>
);

export default ListExperimentsView;
