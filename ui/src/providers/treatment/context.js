import React from "react";

import { useXpApi } from "hooks/useXpApi";

const TreatmentsContext = React.createContext({});

export const TreatmentsContextProvider = ({ projectId, children }) => {
  const [
    {
      data: { data: treatments },
      isLoaded,
    },
  ] = useXpApi(
    `/projects/${projectId}/treatments`,
    {
      query: {
        fields: ["id", "name"],
      },
    },
    { data: [] }
  );

  return (
    <TreatmentsContext.Provider
      value={{
        treatments,
        isLoaded: isLoaded,
      }}>
      {children}
    </TreatmentsContext.Provider>
  );
};

export default TreatmentsContext;
