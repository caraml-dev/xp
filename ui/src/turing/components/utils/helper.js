export const mapProtocolLabel = (protocol, value) => {
  if (protocol === "UPI_V1" && value === "payload") {
    return "prediction context";
  }

  return value;
};
