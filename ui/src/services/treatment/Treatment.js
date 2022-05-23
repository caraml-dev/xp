import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";

const jsonBig = require(`json-bigint`);

export class Treatment {
  constructor() {
    this.name = "";
    this.updated_by = "";
    this.configuration = "";
  }

  static fromJson(json) {
    const clone = cloneDeep(json);
    let obj = merge(new Treatment(), clone);

    obj.configuration = JSON.stringify(obj.configuration);

    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    // Unset updated_by before calling API
    obj.updated_by = "";

    // Format treatment
    obj.configuration = jsonBig.parse(this.configuration);
    delete obj.treatment_template;

    return obj;
  }

  /* stringify returns the Treatment's JSON string representation while handling
     big ints using jsonBig.stringify, that the default JSON.stringify fails to handle.
  */
  stringify() {
    return jsonBig.stringify(this.toJSON());
  }
}
