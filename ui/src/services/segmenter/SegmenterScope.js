export const getSegmenterScope = (scope) => {
  const status = {
    project: {
      label: "Project",
      color: "#07C",
    },
    global: {
      label: "Global",
      color: "#00BFB3",
    },
  };
  return status[scope];
};