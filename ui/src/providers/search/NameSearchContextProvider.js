import React, { useState } from "react";

const NameSearchContext = React.createContext({});

export const NameSearchContextProvider = ({ children }) => {
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
    <NameSearchContext.Provider
      value={{
        getFilter: (name) => filters[name],
        getProcessedFilters: () => filters,
        setFilter: updateFilters,
        isFilterSet: () => Object.keys(filters).length !== 0,
      }}>
      {children}
    </NameSearchContext.Provider>
  );
};

export default NameSearchContext;
