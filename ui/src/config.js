import React from "react";

import objectAssignDeep from "object-assign-deep";

/*
 * In development environment, we set xpApiUrl and mlpApiUrl to unreachable paths
 * so that the requests will be made to the given API servers through proxying
 * (setupProxy.js). This is required to bypass CORS restrictions imposed by the browser.
 * In production, the env vars REACT_APP_XP_API and REACT_APP_MLP_API can either be
 * absolute URLs, or relative to the UI if the API is served from the same host.
 * When the API's origin differs from that of the UI, appropriate CORS policies are expected
 * to be in place on the API server.
 */
const apiConfig = {
  apiTimeout: process.env.REACT_APP_API_TIMEOUT || 5000,
  xpApiUrl:
    process.env.NODE_ENV === "development"
      ? "/api/xp/v1"
      : process.env.REACT_APP_XP_API,
  mlpApiUrl:
    process.env.NODE_ENV === "development"
      ? "/api/mlp"
      : process.env.REACT_APP_MLP_API,
};

const authConfig = {
  oauthClientId: process.env.REACT_APP_OAUTH_CLIENT_ID,
};

export const appConfig = {
  environment: process.env.REACT_APP_ENVIRONMENT || "dev",
  homepage: process.env.REACT_APP_HOMEPAGE || process.env.PUBLIC_URL,
  appIcon: "advancedSettingsApp",
  docsUrl: process.env.REACT_APP_USER_DOCS_URL
    ? JSON.parse(process.env.REACT_APP_USER_DOCS_URL)
    : [{ href: "https://github.com/caraml-dev/xp", label: "XP User Guide" }],
  pagination: {
    defaultPageSize: 10,
  },
  tables: {
    defaultTextSize: "s",
    defaultIconSize: "s",
  },
  datetime: {
    format: "YYYY-MM-DDTHH:mm:ssZ",
    formatNoTz: "YYYY-MM-DDTHH:mm:ss",
    tzOffsetMinutes: 0,
    tz: "UTC",
  },
  // Specifies a set of page templating configurations that will soon be controlled by the mlp-ui package
  // TO-DO: to review if these set of specifications are still needed after the update to mlp-ui
  pageTemplate: {
    // Width of page template
    restrictWidth: "90%",
    // Padding of page template
    paddingSize: "none",
  },
  routeNamePathPrefix: "treatment.configuration."
};

const sentryConfig = {
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: appConfig.environment,
  tags: {
    app: "xp-ui",
  },
};

const buildTimeConfig = {
  apiConfig,
  authConfig,
  appConfig,
  sentryConfig,
};

const ConfigContext = React.createContext({});

export const ConfigProvider = ({ children }) => {
  const runTimeConfig = window.xpConfig;
  const config = objectAssignDeep({}, buildTimeConfig, runTimeConfig);

  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  );
};

export const useConfig = () => React.useContext(ConfigContext);
