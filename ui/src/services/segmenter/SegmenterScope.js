export const getSegmenterScope = (scope) => {
  const status = {
    project: {
      label: "Project",
      color: "primary",
    },
    global: {
      label: "Global",
      color: "secondary",
    },
  };
  return status[scope];
};