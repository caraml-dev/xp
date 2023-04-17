import { useContext } from "react";

import { AuthContext, useApi } from "@caraml-dev/ui-lib";

import { useConfig } from "config";

export const useXpApi = (endpoint, options, result, callImmediately = true) => {
  const { apiConfig } = useConfig();
  const authCtx = useContext(AuthContext);

  return useApi(
    endpoint,
    {
      baseApiUrl: apiConfig.xpApiUrl,
      timeout: apiConfig.apiTimeout,
      parseBigInt: true,
      ...options,
    },
    authCtx,
    result,
    callImmediately
  );
};
