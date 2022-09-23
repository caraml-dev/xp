import React, {useEffect, useMemo, useState} from "react";

import { useXpApi } from "hooks/useXpApi";
import moment from "moment";
import { useConfig } from "../../config";

const ExperimentContext = React.createContext({});

export const ExperimentContextProvider = ({ projectId, children }) => {
  const { appConfig } = useConfig();

  const [isAllExperimentsLoaded, setIsAllExperimentsLoaded] = useState(false);
  const [pageIndex, setPageIndex] = useState(0);
  const [allExperiments, setAllExperiments] = useState([]);

  const { start_time, end_time } = useMemo(
    () => {
      let current_time = moment.utc();
      let start_time = current_time.format(appConfig.datetime.format);
      let end_time = current_time.add(1000, "y").format(appConfig.datetime.format);
      return { start_time, end_time };
    },
    [appConfig]
  );

  const [{ data: { data: experiments, paging }, isLoaded }] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      query: {
        start_time: start_time,
        end_time: end_time,
        page: pageIndex + 1,
        page_size: appConfig.pagination.defaultPageSize,
        status: "active"
      },
    },
    { data: [], paging: { total: 0 } }
  );

  useEffect(() => {
    if (isLoaded) {
      if (!!experiments && !isAllExperimentsLoaded) {
        setAllExperiments((curExperiments) => [...curExperiments, ...experiments]);
      }
      if (paging.pages > paging.page) {
        setPageIndex(paging.page);
      } else {
        setIsAllExperimentsLoaded(true);
      }
    }
  }, [isLoaded, experiments, paging, isAllExperimentsLoaded]);

  return (
    <ExperimentContext.Provider
      value={{
        allExperiments,
        isAllExperimentsLoaded,
      }}>
      {children}
    </ExperimentContext.Provider>
  );
};

export default ExperimentContext;
