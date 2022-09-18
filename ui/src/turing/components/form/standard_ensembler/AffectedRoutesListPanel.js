import React, { useEffect, useMemo, useState } from "react";

import { EuiFlexItem, EuiLoadingChart, EuiTextAlign, EuiInMemoryTable, EuiIcon, EuiTextColor } from "@elastic/eui";

import { Panel } from "components/panel/Panel";
import { useXpApi } from "../../../../hooks/useXpApi";
import moment from "moment";
import { useConfig } from "../../../../config";
import { getExperimentStatus } from "../../../../services/experiment/ExperimentStatus";

export const AffectedRoutesListPanel = ({
  projectId,
  routes,
  routeNamePath
}) => {
  const { appConfig } = useConfig();

  const initRouteToExperimentMappings = routes.reduce((m, r) => {m[r.id] = { running: {}, scheduled: {}}; return m}, {})

  const [isAllExperimentsLoaded, setIsAllExperimentsLoaded] = useState(false);
  const [pageIndex, setPageIndex] = useState(0);
  const [allExperiments, setAllExperiments] = useState([]);
  const [routeToExperimentMappings, setRouteToExperimentMappings] = useState(initRouteToExperimentMappings)

  const { start_time, end_time } = useMemo(
    () => {
      let current_time = moment.utc();
      let start_time = current_time.format(appConfig.datetime.format);
      let end_time = current_time.add(1000, "y").format(appConfig.datetime.format);
      return { start_time, end_time };
      },
    [appConfig]
  );

  const [{ data: { data: experiments, paging }, isLoaded }] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      query: {
        start_time: start_time,
        end_time: end_time,
        page: pageIndex + 1,
        page_size: appConfig.pagination.defaultPageSize,
        status: "active"
      },
    },
    { data: [], paging: { total: 0 } }
  );

  const getRouteName = (config, path) => path.split('.').reduce((obj, key) => obj && obj[key], config);

  // reset loaded experiments if routeNamePath or routes changes
  useEffect(() => {
    if (isAllExperimentsLoaded) {
      let newRouteToExperimentMappings = initRouteToExperimentMappings;
      for (let experiment of allExperiments) {
        for (let treatment of experiment.treatments) {
          let configRouteName = getRouteName(treatment.configuration, routeNamePath);
          if (typeof configRouteName === 'string' && configRouteName in newRouteToExperimentMappings) {
            newRouteToExperimentMappings[configRouteName][getExperimentStatus(experiment).label.toLowerCase()][experiment.id] = experiment;
          }
        }
      }
      setRouteToExperimentMappings(newRouteToExperimentMappings);
    }
  }, [routeNamePath, JSON.stringify(routes), isAllExperimentsLoaded]);

  useEffect(() => {
    if (isLoaded) {
      if (!!experiments && !isAllExperimentsLoaded) {
        setAllExperiments((curExperiments) => [...curExperiments, ...experiments]);
      }
      if (paging.pages > paging.page) {
        setPageIndex(paging.page);
      } else {
        setIsAllExperimentsLoaded(true);
      }
    }
  }, [isLoaded, experiments, paging]);

  const columns = [
    {
      field: "status",
      width: "5px",
      render: (_, item) => {
        const isAssigned = routeToExperimentMappings[item.id] ?
          Object.keys(routeToExperimentMappings[item.id].running).length +
          Object.keys(routeToExperimentMappings[item.id].scheduled).length > 0 : false;
        return (
          <EuiIcon
            type={isAssigned ? "check" : "cross"}
            color={isAssigned ? "success" : "danger"}
            size="m"
            style={{ verticalAlign: "sub" }}
          />
        );
      },
    },
    {
      field: "route_name",
      width: "20%",
      name: "Route Name",
      render: (_, item) => {
        const isAssigned = routeToExperimentMappings[item.id] ?
          Object.keys(routeToExperimentMappings[item.id].running).length +
          Object.keys(routeToExperimentMappings[item.id].scheduled).length > 0 : false;
        return (<EuiTextColor color={isAssigned ? "success" : "danger"}>{item.id}</EuiTextColor>);
      },
    },
    {
      field: "running_experiments",
      width: "35%",
      name: "Running Experiments",
      render: (_, item) => <EuiTextColor>{routeToExperimentMappings[item.id] ? Object.keys(routeToExperimentMappings[item.id].running).length : 0}</EuiTextColor>,
    },
    {
      field: "scheduled_experiments",
      width: "35%",
      name: "Scheduled Experiments",
      render: (_, item) => <EuiTextColor>{routeToExperimentMappings[item.id] ? Object.keys(routeToExperimentMappings[item.id].scheduled).length : 0}</EuiTextColor>,
    },
  ];

  return isAllExperimentsLoaded ? (
    <Panel title={"Affected Routes"}>
      <EuiFlexItem>
        <EuiInMemoryTable
          items={routes}
          columns={columns}
          itemId="id"
          isSelectable={false}
        />
      </EuiFlexItem>
    </Panel>
    ) : (
      <EuiTextAlign textAlign="center">
        <EuiLoadingChart size="xl" mono />
      </EuiTextAlign>
    );
};
