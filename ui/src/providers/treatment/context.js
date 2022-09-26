import React from "react";

import { useXpApi } from "hooks/useXpApi";

const TreatmentContext = React.createContext({});

export const TreatmentContextProvider = ({ projectId, children }) => {
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
    <TreatmentContext.Provider
      value={{
        treatments,
        isLoaded: isLoaded,
      }}>
      {children}
    </TreatmentContext.Provider>
  );
};

export default TreatmentContext;
