export const getSegmenterStatus = (segmenter) => {
  const statusMapping = {
    inactive: {
      label: "Inactive",
      color: "default",
      iconType: "cross",
    },
    active: {
      label: "Active",
      color: "success",
      iconType: "check",
    },
  };

  return statusMapping[segmenter.status];
};
