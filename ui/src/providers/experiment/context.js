import React, { useMemo } from "react";

import { useXpApi } from "hooks/useXpApi";
import moment from "moment";
import { useConfig } from "config";

const ExperimentContext = React.createContext({});

export const ExperimentContextProvider = ({ projectId, children }) => {
  const { appConfig } = useConfig();

  const { start_time, end_time } = useMemo(
    () => {
      let current_time = moment.utc();
      let start_time = current_time.format(appConfig.datetime.format);
      let end_time = current_time.add(1000, "y").format(appConfig.datetime.format);
      return { start_time, end_time };
    },
    [appConfig]
  );

  const [{ data: { data: experiments }, isLoaded }] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      query: {
        start_time: start_time,
        end_time: end_time,
        fields: appConfig.listExperimentFields.experimentContextFields,
        status_friendly: ["running", "scheduled"]
      },
    },
    { data: [] }
  );

  return (
    <ExperimentContext.Provider
      value={{
        experiments,
        isLoaded,
      }}>
      {children}
    </ExperimentContext.Provider>
  );
};

export default ExperimentContext;
