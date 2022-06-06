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
import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import ListTreatmentsTable from "./ListTreatmentsTable";
import TreatmentSearchContext, {
  TreatmentSearchContextProvider,
} from "./search/context";

const ListTreatmentsComponent = ({ projectId, props }) => {
  const { appConfig } = useConfig();
  const [results, setResults] = useState({ items: [], totalItemCount: 0 });
  const [page, setPage] = useState({
    index: 0,
    size: appConfig.pagination.defaultPageSize,
  });

  const { getFilter, getProcessedFilters, setFilter, isFilterSet } = useContext(
    TreatmentSearchContext
  );

  const [{ data, isLoaded, error }] = useXpApi(
    `/projects/${projectId}/treatments`,
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
      { text: "Treatments" },
    ]);
  }, []);

  const onRowClick = (item) => props.navigate(`./${item.id}/details`);

  return (
    <EuiPage paddingSize="none">
      <EuiPageBody paddingSize="m">
        <EuiPageHeader>
          <EuiPageHeaderSection>
            <PageTitle
              title="Treatments"
              postpend={
                isFilterSet() && <EuiBadge color="primary">Filtered</EuiBadge>
              }
            />
          </EuiPageHeaderSection>
          <EuiPageHeaderSection>
            <NavigationMenu curPage={"treatments"} props={props} />
            &emsp;
            <EuiButton size="s" onClick={() => props.navigate("./create")} fill>
              Create Treatment
            </EuiButton>
          </EuiPageHeaderSection>
        </EuiPageHeader>

        <EuiPageContent>
          <EuiFlexItem grow={6}>
            <EuiSearchBar
              query={getFilter("search") || ""}
              box={{
                placeholder: "Search Treatment name",
              }}
              onChange={(text) => {
                setPage({ ...page, index: 0 });
                setFilter("search", text.queryText);
              }}
            />
          </EuiFlexItem>
          <EuiSpacer size="s" />
          <ListTreatmentsTable
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

const ListTreatmentsView = ({ projectId, ...props }) => (
  <TreatmentSearchContextProvider>
    <ListTreatmentsComponent projectId={projectId} props={props} />
  </TreatmentSearchContextProvider>
);

export default ListTreatmentsView;
