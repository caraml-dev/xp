import React from "react";

import { useXpApi } from "hooks/useXpApi";

const SegmentContext = React.createContext({});

export const SegmentContextProvider = ({ projectId, children }) => {
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
    <SegmentContext.Provider
      value={{
        segments,
        isLoaded: isLoaded,
      }}>
      {children}
    </SegmentContext.Provider>
  );
};

export default SegmentContext;
