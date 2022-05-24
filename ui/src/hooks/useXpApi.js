import { useContext } from "react";

import { AuthContext, useApi } from "@gojek/mlp-ui";

import { apiConfig } from "config";

export const useXpApi = (endpoint, options, result, callImmediately = true) => {
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
