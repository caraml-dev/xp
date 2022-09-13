import { Fragment, useContext, useEffect } from "react";

import { EuiSearchBar, EuiSpacer, EuiPanel, EuiPageTemplate } from "@elastic/eui";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";

import { ConfigSection } from "components/config_section/ConfigSection";
import { useXpApi } from "hooks/useXpApi";
import NameSearchContext, {
  NameSearchContextProvider,
} from "providers/search/NameSearchContextProvider";
import { ListSegmentersTable } from "settings/segmenters/list/ListSegmentersTable";

const ListSegmentersComponent = ({ projectId, props }) => {
  const { getFilter, getProcessedFilters, setFilter } =
    useContext(NameSearchContext);

  const [
    {
      data: { data: segmenters },
      isLoaded,
      error,
    },
  ] = useXpApi(
    `/projects/${projectId}/segmenters`,
    {
      query: {
        ...getProcessedFilters(),
      },
    },
    []
  );

  useEffect(() => {
    replaceBreadcrumbs([
      { text: "Experiments", href: ".." },
      { text: "Settings", href: "." },
      { text: "Segmenters" },
    ]);
  });

  const onRowClick = (item) => props.navigate(`./${item.name}/details`);

  return (
    <Fragment>
      <EuiSpacer size="m" />
      <EuiPageTemplate.Section color={"transparent"}>
        <ConfigSection title="Segmenters" iconType="package" />
        <EuiPanel>
          <EuiSearchBar
            query={getFilter("search") || ""}
            box={{
              placeholder: "Search Segmenter name",
            }}
            onChange={(text) => {
              setFilter("search", text.queryText);
            }}
          />
          <EuiSpacer size="s" />
          <ListSegmentersTable
            isLoaded={isLoaded}
            items={segmenters || []}
            error={error}
            onRowClick={onRowClick}
          />
        </EuiPanel>
      </EuiPageTemplate.Section>
    </Fragment>
  );
};

export const ListSegmentersView = ({ projectId, ...props }) => (
  <NameSearchContextProvider>
    <ListSegmentersComponent projectId={projectId} props={props} />
  </NameSearchContextProvider>
);
