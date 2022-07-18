/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

const segmentNameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;
export const segmentConfigSchema = yup.lazy((obj) =>
  yup.object().shape(
    Object.keys(obj).reduce((acc, key) => {
      return {
        ...acc,
        [key]: yup.mixed()
          .when("$segmenterTypes", (segmenterTypes) => {
            switch ((segmenterTypes[key] || "").toUpperCase()) {
              case "BOOL":
                return yup.array(
                  yup
                    .bool()
                    .typeError("Array elements must all be of type: BOOL")
                );
              case "INTEGER":
                return yup.array(
                  yup
                    .number()
                    .integer()
                    .typeError(
                      "Array elements must all be of type: INTEGER"
                    )
                );
              case "REAL":
                return yup.array(
                  yup
                    .number()
                    .typeError("Array elements must all be of type: REAL")
                );
              case "STRING":
                return yup.array(
                  yup
                    .string()
                    .typeError("Array elements must all be of type: STRING")
                );
              default:
                return yup.array(); // Type is unknown for deactivated segmenters
            }
          })
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
              return schema;
            }
          ),
      };
    }, {})
  ));

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
    segment: segmentConfigSchema,
  }),
];

export default schema;
