import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";

export class Settings {
  constructor() {
    this.project_id = 0;
    this.randomization_key = "";

    this.segmenters = {
      names: [],
      variables: {},
    };
    this.enable_s2id_clustering = false;
    this.validation_url = "";
    this.treatment_schema = {
      rules: [],
    };
  }

  static fromJson(json) {
    const clone = cloneDeep(json);
    let obj = merge(new Settings(), clone);
    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    delete obj.created_at;
    delete obj.passkey;
    delete obj.project_id;
    delete obj.updated_at;
    delete obj.username;

    if (obj.treatment_schema.rules.length === 0) delete obj.treatment_schema;
    if (obj.validation_url.length === 0) delete obj.validation_url;

    return obj;
  }

  // stringify returns the Setting's JSON string representation
  stringify() {
    return JSON.stringify(this.toJSON());
  }
}
