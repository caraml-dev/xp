import { formatDate, htmlIdGenerator } from "@elastic/eui/lib/services";
import { get, normalizePath, set } from "@gojek/mlp-ui";
import moment from "moment";

import { appConfig } from "config";

export const makeId = htmlIdGenerator();

export const extractErrors = (validationError) => {
  let errors = {};
  if (validationError.inner) {
    for (let err of validationError.inner) {
      const field = err.path.split(".")[0];
      const path = normalizePath(err.path);
      const fieldsErrors = get(errors, path) || [];
      set(errors, field, [...fieldsErrors, err.message]);
    }
  }
  return errors;
};

export const formatDateCell = (value) => (
  <>
    {formatDate(
      moment(value, appConfig.datetime.format).utcOffset(
        appConfig.datetime.tzOffsetMinutes
      )
    )}
  </>
);

export const makeNewTreatment = (treatment) => ({
  uuid: makeId(),
  template: "",
  // Copy from existing treatment, if passed in
  name: treatment?.name || "",
  traffic: treatment?.traffic || 0,
  configuration: !!treatment?.configuration
    ? JSON.stringify(treatment.configuration)
    : "",
});
