/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

const treatmentNameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;

const schema = [
  yup.object().shape({
    name: yup
      .string()
      .required("Name is required")
      .min(4, "Name should be between 4 and 64 characters")
      .max(64, "Name should be between 4 and 64 characters")
      .matches(
        treatmentNameRegex,
        "Name must begin with an alphanumeric character and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$%&:."
      ),
    configuration: yup
      .string()
      .test(
        "Valid JSON",
        "Configuration should be a valid JSON dictionary",
        (item) => {
          try {
            var parsedItem = JSON.parse(item);
            if (typeof parsedItem != "object" || Array.isArray(parsedItem)) {
              return false;
            }
          } catch (e) {
            return false;
          }
          return true;
        }
      ),
  }),
];

export default schema;
