export const getSegmenterStatus = (segmenter) => {
  const statusMapping = {
    inactive: {
      label: "inactive",
      color: "#6A717D",
      healthColor: "subdued",
      iconType: "cross",
    },
    active: {
      label: "active",
      color: "#017D73",
      healthColor: "success",
      iconType: "check",
    },
  };

  return statusMapping[segmenter.status];
};
