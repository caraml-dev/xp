/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

const segmentNameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;

const schema = [
  yup.object().shape({
    name: yup
      .string()
      .required("Name is required")
      .min(4, "Name should be between 4 and 64 characters")
      .max(64, "Name should be between 4 and 64 characters")
      .matches(
        segmentNameRegex,
        "Name must begin with an alphanumeric character and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$%&:."
      ),
    segment: yup.lazy((obj) =>
      yup.object().shape(
        Object.keys(obj).reduce((acc, key) => {
          return {
            ...acc,
            [key]: yup
              .array()
              .when(
                "$requiredSegmenterNames",
                (requiredSegmenterNames, schema) => {
                  if (requiredSegmenterNames.includes(key)) {
                    return schema
                      .required(`Segmenter ${key} is required`)
                      .min(
                        1,
                        `Segmenter ${key} should at least have 1 valid value`
                      );
                  }
                }
              ),
          };
        }, {})
      )
    ),
  }),
];

export default schema;
