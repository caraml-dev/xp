import React, { useState } from "react";

const TreatmentSearchContext = React.createContext({});

export const TreatmentSearchContextProvider = ({ children }) => {
  const [filters, setFilters] = useState({});

  const updateFilters = (name, value) => {
    let newFilters = { ...filters };
    if (!value || (Array.isArray(value) && !value.length)) {
      delete newFilters[name];
    } else {
      newFilters[name] = value;
    }

    setFilters(newFilters);
  };

  return (
    <TreatmentSearchContext.Provider
      value={{
        getFilter: (name) => filters[name],
        getProcessedFilters: () => filters,
        setFilter: updateFilters,
        isFilterSet: () => Object.keys(filters).length !== 0,
      }}>
      {children}
    </TreatmentSearchContext.Provider>
  );
};

export default TreatmentSearchContext;
