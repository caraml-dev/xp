/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

const settingsRandomizationKeyRegex = /^[a-zA-Z0-9-_]*$/;

var JSONbig = require("json-bigint");

function validateSegmenterSelection(items) {
  const dependencyMap = this.options.context["dependencyMap"];
  const accumulatedErrors = [];
  for (const [key, value] of Object.entries(dependencyMap)) {
    value.forEach((dependentSegmenter) => {
      if (items.includes(dependentSegmenter) && !items.includes(key)) {
        accumulatedErrors.push(
          this.createError({
            path: `${this.path}`,
            message: `${key} is required for ${dependentSegmenter}`,
          })
        );
      }
    });
  }
  if (accumulatedErrors.length) {
    return new yup.ValidationError(accumulatedErrors);
  }
  return true;
}

const ruleSchema = yup.object().shape({
  name: yup.string().required("Name is required"),
  predicate: yup.string().required("Predicate is required"),
});

const validateTreatmentRuleNames = function (items) {
  const uniqueNamesMap = items.reduce((acc, item) => {
    const current = item.name in acc ? acc[item.name] : 0;
    // If name is set, increment the count
    return !!item.name ? { ...acc, [item.name]: current + 1 } : acc;
  }, {});
  const errors = [];
  items.forEach((item, idx) => {
    if (!!item.name && uniqueNamesMap[item.name] > 1) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].name`,
          message: "Rules names in a project should be unique",
        })
      );
    }
  });
  return !!errors.length ? new yup.ValidationError(errors) : true;
};

const schema = [
  //RandomizationStep
  yup.object().shape({
    randomization_key: yup
      .string()
      .required("Randomization Key is required")
      .min(4, "Randomization Key should be between 4 and 64 characters")
      .max(64, "Randomization Key should be between 4 and 64 characters")
      .matches(
        settingsRandomizationKeyRegex,
        "Randomization Key can only contain letters a-z (or capitalized), numbers 0-9, the hyphen - symbol and the underscore _ symbol"
      ),
  }),
  // SegmentersStep
  yup.object().shape({
    dependencyMap: yup.object(),
    segmenters: yup.object().shape({
      names: yup
        .array()
        .test("segmenter-dependencies", validateSegmenterSelection),
      variables: yup.object().when("names", ([names], schema) => {
        const shape = names.reduce((acc, name) => {
          acc[name] = yup
            .array()
            .of(yup.string())
            .required("Experiment variable must be selected");
          return acc;
        }, {});
        return schema.shape(shape);
      }),
    }),
  }),
  // ExternalValidationStep
  yup.object().shape({
    validation_url: yup.string().url(),
  }),
  // TreatmentValidationStep
  yup.object().shape({
    treatment_schema: yup.object().shape({
      rules: yup
        .array()
        .of(ruleSchema)
        .test("unique name", validateTreatmentRuleNames),
    }),
  }),
  // Playground treatment configuration validation
  yup
    .string()
    .test(
      "Valid JSON",
      "Configuration should be a valid JSON dictionary",
      (item) => {
        try {
          var parsedItem = JSONbig.parse(item);
          if (typeof parsedItem != "object" || Array.isArray(parsedItem)) {
            return false;
          }
        } catch (e) {
          return false;
        }
        return true;
      }
    ),
];

export default schema;
