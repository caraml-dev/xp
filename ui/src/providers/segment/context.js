import React from "react";

import { useXpApi } from "hooks/useXpApi";

const SegmentsContext = React.createContext({});

export const SegmentsContextProvider = ({ projectId, children }) => {
  const [
    {
      data: { data: segments },
      isLoaded,
    },
  ] = useXpApi(
    `/projects/${projectId}/segments`,
    {
      query: {
        fields: ["id", "name"],
      },
    },
    { data: [] }
  );

  return (
    <SegmentsContext.Provider
      value={{
        segments,
        isLoaded: isLoaded,
      }}>
      {children}
    </SegmentsContext.Provider>
  );
};

export default SegmentsContext;
