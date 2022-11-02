export const mapProtocolLabel = (protocol, value) => {
  if (protocol === "UPI_V1" && value === "payload") {
    return "prediction context";
  }

  return value;
};

export const capitalizeFirstLetter = (string) => {
  return string.replace(/(^\w|\s\w)/g, (m) => m.toUpperCase());
};
