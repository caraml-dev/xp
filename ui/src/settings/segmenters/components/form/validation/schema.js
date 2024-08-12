/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

import { typeOptions } from "settings/segmenters/components/typeOptions";

const nameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;
const nameRegexDescription =
  "Name must begin with an alphanumeric character and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$%&:.";
const typeOptionsValues = typeOptions.map((e) => e.value);

const segmenterValueSchema = yup.mixed();

const segmenterValuesSchema = yup
  .array(segmenterValueSchema)
  .min(1, "Allowed values array cannot be empty")
  .test(
    "Valid JSON array",
    "Segmenter values must be a valid array of values of the same type",
    (array) => {
      var baseType = typeof array[0];
      for (const item of array) {
        if (typeof item !== baseType) {
          return false;
        }
      }
      return true;
    }
  )
  .typeError(
    "Segmenter values must be a valid array of string, bool, integer or real values"
  );

const preRequisiteSchema = yup.object().shape({
  segmenter_name: yup.string(),
  segmenter_values: segmenterValuesSchema,
});

const optionsSchema = yup
  .string()
  .test(
    "Valid JSON",
    "Option values should be a valid JSON dictionary",
    (item) => {
      if (item !== "") {
        try {
          var parsedItem = JSON.parse(item);
          if (typeof parsedItem != "object" || Array.isArray(parsedItem)) {
            return false;
          }
        } catch (e) {
          return false;
        }
      }
      return true;
    }
  );

const validateArrayString = (arraySchema, arrayName) => {
  return yup
    .string()
    .test(
      `${arrayName} valid JSON array`,
      `${arrayName} must be a valid JSON array`,
      (array) => {
        if (array !== "") {
          try {
            var parsedArray = JSON.parse(array);
            if (typeof parsedArray != "array" && !Array.isArray(parsedArray)) {
              return false;
            }
            return arraySchema
              .validateSync(parsedArray);
          } catch (e) {
            return false;
          }
          
        }
        return true;
      }
    )
}

const constraintSchema = yup.object().shape({
  pre_requisites: validateArrayString(yup.array(preRequisiteSchema), "Pre-requisites")
    .typeError(
      "Constraint pre-requisites must be a valid array of pre-requisite objects"
    ),
  allowed_values: validateArrayString(segmenterValuesSchema, "Allowed values"),
  options: optionsSchema,
});

const schema = [
  yup.object().shape({
    name: yup
      .string()
      .required("Name is required")
      .min(4, "Name should be between 4 and 64 characters")
      .max(64, "Name should be between 4 and 64 characters")
      .matches(nameRegex, nameRegexDescription),
    type: yup
      .string()
      .required("Segmenter Value Type should be selected")
      .oneOf(
        typeOptionsValues,
        "Valid Segmenter Value Type should be selected"
      ),
    description: yup.string(),
    required: yup.bool(),
    multi_valued: yup.bool(),
  }),
  yup.object().shape({
    options: optionsSchema,
  }),
  yup.object().shape({
    constraints: yup.array(constraintSchema),
  }),
];

export default schema;
