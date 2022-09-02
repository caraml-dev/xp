export const getSegmenterStatus = (segmenter) => {
  const statusMapping = {
    inactive: {
      label: "Inactive",
      color: "#6A717D",
      healthColor: "subdued",
      iconType: "cross",
    },
    active: {
      label: "Active",
      color: "#00BFB3",
      healthColor: "success",
      iconType: "check",
    },
  };

  return statusMapping[segmenter.status];
};
