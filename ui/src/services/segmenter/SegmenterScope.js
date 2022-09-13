export const getSegmenterScope = (scope) => {
  const status = {
    project: {
      label: "Project",
      color: "primary",
    },
    global: {
      label: "Global",
      color: "success",
    },
  };
  return status[scope];
};