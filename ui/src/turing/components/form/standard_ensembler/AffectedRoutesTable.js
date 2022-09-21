import React, { useEffect, useMemo, useState } from "react";

import {
  EuiFlexItem,
  EuiLoadingChart,
  EuiTextAlign,
  EuiInMemoryTable,
  EuiIcon,
  EuiTextColor,
} from "@elastic/eui";

import { useXpApi } from "hooks/useXpApi";
import moment from "moment";
import { useConfig } from "config";
import { getExperimentStatus } from "services/experiment/ExperimentStatus";
import { AffectedExperimentsContextMenu } from "./AffectedExperimentsContextMenu";

export const AffectedRoutesTable = ({
  projectId,
  routes,
  treatmentConfigRouteNamePath,
}) => {
  const { appConfig } = useConfig();

  const [isButtonPopoverOpen, setIsButtonPopoverOpen] = useState(routes.reduce((m, r) => {m[r.id] = {running: false, scheduled: false}; return m}, {}));
  const [isAllExperimentsLoaded, setIsAllExperimentsLoaded] = useState(false);
  const [pageIndex, setPageIndex] = useState(0);
  const [allExperiments, setAllExperiments] = useState([]);
  const [routeToExperimentMappings, setRouteToExperimentMappings] = useState(routes.reduce((m, r) => {m[r.id] = {running: {}, scheduled: {}}; return m}, {}));

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

  // this stringified value of routes below allows the React effect below to mimic a deep comparison when changes to the
  // array routes are made
  const stringifiedRoutes = JSON.stringify(routes)

  // reset loaded routeToExperimentMappings if treatmentConfigRouteNamePath or routes changes
  useEffect(() => {
    if (isAllExperimentsLoaded) {
      let newRouteToExperimentMappings = routes.reduce((m, r) => {m[r.id] = {running: {}, scheduled: {}}; return m}, {});
      for (let experiment of allExperiments) {
        for (let treatment of experiment.treatments) {
          let configRouteName = getRouteName(treatment.configuration, treatmentConfigRouteNamePath);
          if (typeof configRouteName === 'string' && configRouteName in newRouteToExperimentMappings) {
            newRouteToExperimentMappings[configRouteName][getExperimentStatus(experiment).label.toLowerCase()][experiment.id] = experiment;
          }
        }
      }
      setRouteToExperimentMappings(newRouteToExperimentMappings);
      setIsButtonPopoverOpen(routes.reduce((m, r) => {m[r.id] = {running: false, scheduled: false}; return m}, {}));
    }
  }, [treatmentConfigRouteNamePath, stringifiedRoutes, routes, isAllExperimentsLoaded, allExperiments]);

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
  }, [isLoaded, experiments, paging, isAllExperimentsLoaded]);

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
            size={"m"}
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
      render: (_, item) => (
        <AffectedExperimentsContextMenu
          item={item}
          projectId={projectId}
          routeToExperimentMappings={routeToExperimentMappings}
          isButtonPopoverOpen={isButtonPopoverOpen}
          setIsButtonPopoverOpen={setIsButtonPopoverOpen}
          experimentStatus={"running"}
        />
      ),
    },
    {
      field: "scheduled_experiments",
      width: "35%",
      name: "Scheduled Experiments",
      render: (_, item) => (
        <AffectedExperimentsContextMenu
          item={item}
          projectId={projectId}
          routeToExperimentMappings={routeToExperimentMappings}
          isButtonPopoverOpen={isButtonPopoverOpen}
          setIsButtonPopoverOpen={setIsButtonPopoverOpen}
          experimentStatus={"scheduled"}
        />
      )
    },
  ];

  return isAllExperimentsLoaded ? (
    <EuiFlexItem>
      <EuiInMemoryTable
        items={routes.filter((r) => r.id !== "")}
        columns={columns}
        itemId={"id"}
        isSelectable={false}
      />
    </EuiFlexItem>
    ) : (
      <EuiTextAlign textAlign={"center"}>
        <EuiLoadingChart size={"xl"} mono />
      </EuiTextAlign>
    );
};
