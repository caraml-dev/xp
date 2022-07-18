var jsonBig = require(`json-bigint`);

export class Segment { }

export const parseSegmenterValue = (value, type) => {
  let parsedValue;
  switch (type.toUpperCase()) {
    case "BOOL":
      parsedValue = value.toLowerCase() === "true";
      break;
    case "INTEGER":
      parsedValue = jsonBig.parse(value);
      break;
    case "REAL":
      parsedValue = Number(value);
      break;
    default:
      parsedValue = value;
  }
  return parsedValue;
};

export const stringifySegmenterValue = (item) => {
  return typeof item === "string" ? item : jsonBig.stringify(item);
};
