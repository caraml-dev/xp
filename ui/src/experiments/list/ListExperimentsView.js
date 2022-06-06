import React, { useContext, useEffect, useState } from "react";

import {
  EuiBadge,
  EuiButton,
  EuiFlexGroup,
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
import classNames from "classnames";

import { NavigationMenu } from "components/page/NavigationMenu";
import { PageTitle } from "components/page/PageTitle";
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListExperimentsTable from "./ListExperimentsTable";
import ExperimentSearchContext, {
  ExperimentSearchContextProvider,
} from "./search/context";
import SearchExperimentsPanel from "./search/SearchExperimentsPanel";

import "./ListExperimentsView.scss";

const ListExperimentsComponent = ({ projectId, props }) => {
  const { appConfig } = useConfig();
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
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

  const onRowClick = (item) => props.navigate(`./${item.id}/details`);

  return (
    <EuiPage
      paddingSize="none"
      className={classNames({ pageWithLeftSidebar: isSearchPanelVisible })}>
      {isSearchPanelVisible && (
        <SearchExperimentsPanel
          onChange={() => setPage({ ...page, index: 0 })}
          onClose={setIsSearchPanelVisible}
          projectId={projectId}
        />
      )}

      <EuiPageBody paddingSize="m">
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle
              title="Experiments"
              postpend={
                isFilterSet() && <EuiBadge color="primary">Filtered</EuiBadge>
              }
            />
          </EuiPageHeaderSection>
          <EuiPageHeaderSection>
            <NavigationMenu curPage={"experiments"} props={props} />
            &emsp;
            <EuiButton size="s" onClick={() => props.navigate("./create")} fill>
              Create Experiment
            </EuiButton>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContent>
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
            props={props}
          />
        </EuiPageContent>
      </EuiPageBody>
    </EuiPage>
  );
};

const ListExperimentsView = ({ projectId, ...props }) => (
  <ExperimentSearchContextProvider>
    <ListExperimentsComponent projectId={projectId} props={props} />
  </ExperimentSearchContextProvider>
);

export default ListExperimentsView;
