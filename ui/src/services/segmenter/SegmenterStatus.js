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
      color: "#017D73",
      healthColor: "success",
      iconType: "check",
    },
  };

  return statusMapping[segmenter.status];
};
