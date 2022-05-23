import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";

import {
  treatment_schema,
  validation_url,
} from "settings/components/playground_flyout/typeOptions";

var JSONbig = require("json-bigint");

export class ValidateEntityRequest {
  constructor() {
    this.data = "";
    this.validation_url = "";
    this.treatment_schema = {
      rules: [],
    };
  }

  static fromSettings(settings) {
    let obj = new ValidateEntityRequest();
    obj.validation_url = settings.validation_url;
    obj.treatment_schema = settings.treatment_schema;
    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    if (obj.treatment_schema.rules.length === 0) delete obj.treatment_schema;
    if (obj.validation_url.length === 0) delete obj.validation_url;

    return obj;
  }

  setupRequest(validationOption) {
    const clone = cloneDeep(this);
    if (validationOption !== validation_url) {
      clone.validation_url = "";
    }
    if (validationOption !== treatment_schema) {
      clone.treatment_schema.rules = [];
    }
    return clone;
  }

  // stringify returns the ValidateEntityRequest's JSON string representation
  stringify() {
    return JSONbig.stringify(this.toJSON());
  }
}
