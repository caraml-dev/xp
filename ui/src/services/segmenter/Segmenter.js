import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";

const jsonBig = require(`json-bigint`);

export class Segmenter {
  constructor() {
    this.name = "";
    this.type = "";
    this.description = "";
    this.required = false;
    this.multi_valued = false;

    this.options = "";

    this.constraints = [];
  }

  static fromJson(json) {
    const clone = cloneDeep(json);
    let obj = merge(new Segmenter(""), clone);

    obj.options =
      obj?.options != null && Object.keys(obj.options).length !== 0
        ? JSON.stringify(obj?.options)
        : "";
    obj.constraints = obj.constraints?.map(newConstraint) || [];
    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    // Format options
    obj.options !== ""
      ? (obj.options = jsonBig.parse(this.options))
      : delete obj.options;

    // Format contraints
    obj.constraints = obj.constraints?.map((constraint) => {
      constraint.pre_requisites = jsonBig.parse(constraint.pre_requisites);
      constraint.allowed_values = jsonBig.parse(constraint.allowed_values);
      constraint.options !== ""
        ? (constraint.options = jsonBig.parse(constraint.options))
        : delete constraint.options;
      return constraint;
    });
    if (obj.constraints != null && obj.constraints.length === 0)
      delete obj.constraints;

    return obj;
  }

  /* stringify returns the Experiment's JSON string representation while handling
     big ints using jsonBig.stringify, that the default JSON.stringify fails to handle.
  */
  stringify() {
    return jsonBig.stringify(this.toJSON());
  }
}

export const newConstraint = (segmenter) => {
  if (!!segmenter) {
    return {
      pre_requisites:
        segmenter?.pre_requisites != null
          ? JSON.stringify(segmenter?.pre_requisites)
          : "[]",
      allowed_values:
        segmenter?.allowed_values != null
          ? JSON.stringify(segmenter?.allowed_values)
          : "[]",
      options:
        segmenter?.options != null &&
          Object.keys(segmenter.options).length !== 0
          ? JSON.stringify(segmenter?.options)
          : "",
    };
  } else {
    return {
      pre_requisites: "",
      allowed_values: "",
      options: "",
    };
  }
};
