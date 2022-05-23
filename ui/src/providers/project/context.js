import React, { useEffect, useState } from "react";

import { useXpApi } from "hooks/useXpApi";

const ProjectContext = React.createContext({});

export const ProjectContextProvider = ({ children }) => {
  const [projectSegmentersMap, setProjectSegmentersMap] = useState({});
  const [
    {
      data: { data: projects },
      isLoaded,
    },
  ] = useXpApi(`/projects`, {}, []);

  useEffect(() => {
    if (isLoaded && projects) {
      let projSegmentersMap = {};
      projects.forEach((p) => {
        projSegmentersMap[p.id] = p.segmenters;
      });
      setProjectSegmentersMap(projSegmentersMap);
    }
  }, [projects, isLoaded]);

  return (
    <ProjectContext.Provider
      value={{
        isProjectOnboarded: (projectId) => {
          if (!!projects) {
            const projectIds = projects.map((project) => project.id.toString());
            return projectIds.includes(projectId);
          }
          return false;
        },
        isLoaded,
        projectSegmentersMap,
      }}>
      {children}
    </ProjectContext.Provider>
  );
};

export default ProjectContext;
