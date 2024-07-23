/* eslint-disable no-template-curly-in-string */
import * as yup from "yup";

import {
  experimentStatuses,
  experimentTiers,
  experimentTypes,
} from "experiments/components/typeOptions";
import { segmentConfigSchema } from "segments/components/form/validation/schema";

const nameRegex = /^[A-Za-z\d][\w\d \-()#$%&:.]*[\w\d\-()#$%&:.]$/;
const nameRegexDescription =
  "Name must begin with an alphanumeric character and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$%&:.";
const experimentStatusValues = experimentStatuses.map((e) => e.value);
const experimentTypeValues = experimentTypes.map((e) => e.value);
const experimentTierValues = experimentTiers.map((e) => e.value);

// Note: The validation functions below are specifically not defined using the
// arrow format, for access to `this` from the invocation context:
// https://stackoverflow.com/a/33308151

const validateABTreatmentTraffic = function(items) {
  const sum = items.reduce(
    (total, e) => total + (!!e.traffic ? e.traffic : 0),
    0
  );
  const errors = [];
  items.forEach((item, idx) => {
    if (!item.traffic) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].traffic`,
          message: "Traffic cannot be 0 for A/B experiments",
        })
      );
    }
    if (sum !== 100) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].traffic`,
          message: "Traffic should sum up to 100",
        })
      );
    }
  });
  return !!errors.length ? new yup.ValidationError(errors) : true;
};

const validateSwitchbackTreatmentTraffic = function(items) {
  const sum = items.reduce(
    (total, e) => total + (!!e.traffic ? e.traffic : 0),
    0
  );
  const withTraffic = items.filter((e) => !!e.traffic);
  const errors = [];
  items.forEach((_, idx) => {
    if (withTraffic.length !== 0 && withTraffic.length !== items.length) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].traffic`,
          message:
            "Traffic should be set for all treatments or none in Switchback experiments",
        })
      );
    }
    if (sum !== 0 && sum !== 100) {
      errors.push(
        this.createError({
          path: `${this.path}[${idx}].traffic`,
          message: "Traffic should sum up to 0 or 100",
        })
      );
    }
  });
  return !!errors.length ? new yup.ValidationError(errors) : true;
};

const validateTreatmentNames = function(items) {
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
          message: "Treatment names in an experiment should be unique",
        })
      );
    }
  });
  return !!errors.length ? new yup.ValidationError(errors) : true;
};

const treatmentTrafficSchema = yup
  .number()
  .integer("Traffic percentage must be an integer")
  .min(0, "Traffic percentage cannot be lower than 1")
  .max(100, "Traffic percentage cannot be greater than 100")
  .optional();

const treatmentSchema = yup.object().shape({
  name: yup
    .string()
    .required("Name is required")
    .min(4, "Name should be between 4 and 64 characters")
    .max(64, "Name should be between 4 and 64 characters")
    .matches(nameRegex, nameRegexDescription),
  traffic: treatmentTrafficSchema,
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
      .required("Experiment Type should be selected")
      .oneOf(experimentTypeValues, "Valid Experiment Type should be selected"),
    interval: yup.string().when("type", ([type], schema) => {
      return type === "Switchback"
        ? yup
          .number()
          .required("Interval is required for Switchback experiments")
          .integer("Interval must be an integer")
          .min(5, "Interval cannot be lower than 5 minutes")
        : schema;
    }),
    tier: yup
      .string()
      .required("Experiment Tier should be selected")
      .oneOf(experimentTierValues, "Valid Experiment Tier should be selected"),
    status: yup
      .string()
      .required("Experiment Status should be selected")
      .oneOf(
        experimentStatusValues,
        "Valid Experiment Status should be selected"
      ),
    start_time: yup.date().required("Start Time is required").min(new Date()),
    end_time: yup
      .date()
      .required("End Time is required")
      .when("start_time", ([startTime], schema) => {
        return startTime ? schema.min(startTime) : schema;
      }),
  }),
  yup.object().shape({
    segment: segmentConfigSchema,
  }),
  yup.object().shape({
    treatments: yup
      .array()
      .of(treatmentSchema)
      .when("type", ([type], schema) => {
        switch (type) {
          case "A/B":
            return schema
              .test("traffic-and-sum-100", validateABTreatmentTraffic)
              .test("unique-treatment-names", validateTreatmentNames);
          case "Switchback":
            return schema
              .test("traffic-sum-0-or-100", validateSwitchbackTreatmentTraffic)
              .test("unique-treatment-names", validateTreatmentNames);
          default:
            return schema;
        }
      }),
  }),
];

export default schema;
