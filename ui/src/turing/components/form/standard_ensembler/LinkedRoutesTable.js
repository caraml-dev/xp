import React, { useContext, useEffect, useState } from "react";

import {
  EuiFlexItem,
  EuiLoadingChart,
  EuiTextAlign,
  EuiInMemoryTable,
  EuiIcon,
  EuiTextColor,
} from "@elastic/eui";

import { LinkedExperimentsContextMenu } from "./LinkedExperimentsContextMenu";
import ExperimentContext from "providers/experiment/context";

export const LinkedRoutesTable = ({
  projectId,
  routes,
  treatmentConfigRouteNamePath,
}) => {
  const { allExperiments, isAllExperimentsLoaded } = useContext(ExperimentContext)

  const [routeToExperimentMappings, setRouteToExperimentMappings] = useState(routes.reduce((m, r) => {m[r.id] = {running: {}, scheduled: {}}; return m}, {}));

  const getRouteName = (config, path) => path.split('.').reduce((obj, key) => obj && obj[key], config);

  // this stringified value of routes below allows the React effect below to mimic a deep comparison when changes to the
  // array routes are made
  const stringifiedRoutes = routes.map(e => e.id).join();

  // reset loaded routeToExperimentMappings if treatmentConfigRouteNamePath or routes changes
  useEffect(() => {
    if (isAllExperimentsLoaded) {
      let newRouteToExperimentMappings = routes.reduce((m, r) => {m[r.id] = {running: {}, scheduled: {}}; return m}, {});
      for (let experiment of allExperiments) {
        for (let treatment of experiment.treatments) {
          let configRouteName = getRouteName(treatment.configuration, treatmentConfigRouteNamePath);
          if (typeof configRouteName === 'string' && configRouteName in newRouteToExperimentMappings) {
            newRouteToExperimentMappings[configRouteName][experiment.status_friendly][experiment.id] = experiment;
          }
        }
      }
      setRouteToExperimentMappings(newRouteToExperimentMappings);
    }
  }, [treatmentConfigRouteNamePath, stringifiedRoutes, routes, isAllExperimentsLoaded, allExperiments]);

  const columns = [
    {
      field: "id",
      width: "5px",
      render: (id) => {
        const isAssigned = routeToExperimentMappings[id] ?
          Object.keys(routeToExperimentMappings[id].running).length +
          Object.keys(routeToExperimentMappings[id].scheduled).length > 0 : false;
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
      field: "id",
      width: "20%",
      name: "Route Name",
      render: (id) => {
        const isAssigned = routeToExperimentMappings[id] ?
          Object.keys(routeToExperimentMappings[id].running).length +
          Object.keys(routeToExperimentMappings[id].scheduled).length > 0 : false;
        return (<EuiTextColor color={isAssigned ? "success" : "danger"}>{id}</EuiTextColor>);
      },
    },
    {
      field: "id",
      width: "35%",
      name: "Running Experiments",
      render: (id) => routeToExperimentMappings[id] && (
        <LinkedExperimentsContextMenu
          projectId={projectId}
          linkedExperiments={routeToExperimentMappings[id]}
          experimentStatus={"running"}
        />
      ),
    },
    {
      field: "id",
      width: "35%",
      name: "Scheduled Experiments",
      render: (id) => routeToExperimentMappings[id] && (
        <LinkedExperimentsContextMenu
          projectId={projectId}
          linkedExperiments={routeToExperimentMappings[id]}
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
