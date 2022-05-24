import React from "react";

import { useXpApi } from "hooks/useXpApi";

const SettingsContext = React.createContext({});

export const SettingsContextProvider = ({ projectId, children }) => {
  const [
    {
      data: { data: settings },
      isLoaded: isSettingsLoaded,
    },
  ] = useXpApi(`/projects/${projectId}/settings`, {}, {});

  const [
    {
      data: { data: variables },
      isLoaded: isVariablesLoaded,
    },
  ] = useXpApi(`/projects/${projectId}/experiment-variables`, {}, []);

  return (
    <SettingsContext.Provider
      value={{
        settings,
        variables,
        isLoaded: (name) => {
          switch (name) {
            case "settings":
              return isSettingsLoaded;
            case "variables":
              return isVariablesLoaded;
            default:
              return false;
          }
        },
      }}>
      {children}
    </SettingsContext.Provider>
  );
};

export default SettingsContext;
