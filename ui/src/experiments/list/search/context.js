import React, { useState } from "react";

const ExperimentSearchContext = React.createContext({});

export const ExperimentSearchContextProvider = ({ children }) => {
  const [filters, setFilters] = useState({});

  const getProcessedFilters = () => {
    const newFilters = { ...filters };
    /*
      Valid selections include:
      - Have both start_time and end_time or NONE;
      - start_time < end_time
    */
    if (
      !(!!newFilters["start_time"] && !!newFilters["end_time"]) ||
      (!!newFilters["start_time"] &&
        !!newFilters["end_time"] &&
        newFilters["end_time"] < newFilters["start_time"])
    ) {
      delete newFilters["start_time"];
      delete newFilters["end_time"];
    }
    // Apply hour, min, sec selections 00:00:00 and 23:59:59 to start_time and end_time respectively
    if (newFilters["start_time"]) {
      newFilters["start_time"] = newFilters["start_time"]
        .startOf("day")
        .format("YYYY-MM-DDTHH:mm:ssZ");
    }
    if (newFilters["end_time"]) {
      newFilters["end_time"] = newFilters["end_time"]
        .endOf("day")
        .format("YYYY-MM-DDTHH:mm:ssZ");
    }
    return newFilters;
  };

  const updateFilters = (name, value, filtersToDelete = []) => {
    let newFilters = { ...filters };
    if (!value || (Array.isArray(value) && !value.length)) {
      delete newFilters[name];
    } else {
      newFilters[name] = value;
    }
    if (filtersToDelete) {
      filtersToDelete.forEach((filter) => {
        // Reset filters
        delete newFilters[filter];
      });
    }

    setFilters(newFilters);
  };

  return (
    <ExperimentSearchContext.Provider
      value={{
        getFilter: (name) => filters[name],
        getFilters: () => ({ ...filters }),
        getProcessedFilters,
        setFilter: updateFilters,
        clearFilters: () => setFilters({}),
        isFilterSet: () => Object.keys(getProcessedFilters()).length !== 0,
      }}>
      {children}
    </ExperimentSearchContext.Provider>
  );
};

export default ExperimentSearchContext;
