import * as yup from "yup";

const schema = yup.lazy((obj) => {
  if (!!obj.start_time || !!obj.end_time) {
    return yup.object().shape({
      start_time: yup.date().required("End time is set, Start time required"),
      end_time: yup.mixed().when("start_time", (startTime, schema) => {
        return !!startTime
          ? yup
              .date()
              .required("Start time is set, End time required")
              .min(startTime)
          : schema;
      }),
    });
  }
  return yup.object();
});

export default schema;
