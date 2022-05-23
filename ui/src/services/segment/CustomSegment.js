import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";

import { Segment } from "services/experiment/Segment";

const jsonBig = require(`json-bigint`);

export class CustomSegment {
  constructor() {
    this.name = "";
    this.updated_by = "";
    this.segment = new Segment();
  }

  static fromJson(json) {
    const clone = cloneDeep(json);
    let obj = merge(new CustomSegment(""), clone);

    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    // Unset updated_by before calling API
    obj.updated_by = "";

    // Format segment
    delete obj.segment_template;

    return obj;
  }

  /* stringify returns the Segment's JSON string representation while handling
     big ints using jsonBig.stringify, that the default JSON.stringify fails to handle.
  */
  stringify() {
    return jsonBig.stringify(this.toJSON());
  }
}
