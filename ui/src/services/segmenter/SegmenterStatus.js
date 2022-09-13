export const getSegmenterStatus = (segmenter) => {
  const statusMapping = {
    inactive: {
      label: "Inactive",
      color: "subdued",
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
