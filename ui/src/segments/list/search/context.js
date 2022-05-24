import React, { useState } from "react";

const SegmentSearchContext = React.createContext({});

export const SegmentSearchContextProvider = ({ children }) => {
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
    <SegmentSearchContext.Provider
      value={{
        getFilter: (name) => filters[name],
        getProcessedFilters: () => filters,
        setFilter: updateFilters,
        isFilterSet: () => Object.keys(filters).length !== 0,
      }}>
      {children}
    </SegmentSearchContext.Provider>
  );
};

export default SegmentSearchContext;
