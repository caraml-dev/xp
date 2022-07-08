import cloneDeep from "lodash/cloneDeep";
import merge from "lodash/merge";
import moment from "moment";

import { appConfig } from "config";
import { makeNewTreatment } from "utils/helpers";

import { Segment } from "./Segment";

const jsonBig = require(`json-bigint`);

export class Experiment {
  constructor() {
    this.id = 0;
    this.description = "";
    this.name = "";
    this.updated_by = "";

    this.start_time = undefined;
    this.end_time = undefined;
    this.type = "";
    this.tier = "";
    this.status = "";

    this.segment = new Segment();
    this.segment_template = "";
    this.treatments = [];
    this.interval = 0;
  }

  static fromJson(json) {
    const clone = cloneDeep(json);
    let obj = merge(new Experiment(""), clone);

    if (!obj.interval) {
      // Interval could be null or undefined for Switchbacks. Make it 0.
      obj.interval = 0;
    }

    obj.start_time = moment(
      obj.start_time,
      appConfig.datetime.format
    ).utcOffset(appConfig.datetime.tzOffsetMinutes);
    obj.end_time = moment(obj.end_time, appConfig.datetime.format).utcOffset(
      appConfig.datetime.tzOffsetMinutes
    );

    obj.treatments = obj.treatments.map(makeNewTreatment);
    return obj;
  }

  toJSON() {
    const clone = cloneDeep(this);
    let obj = merge({}, clone);

    // Unset updated_by before calling API
    obj.updated_by = "";

    // Format treatments
    obj["treatments"] = obj["treatments"].map((treatment) => {
      delete treatment.uuid;
      delete treatment.template;
      treatment.configuration = jsonBig.parse(treatment.configuration);
      return treatment;
    });

    // Format segment
    delete obj.segment_template;

    // Format start and end times
    obj = {
      ...obj,
      start_time: this.start_time.format(appConfig.datetime.format),
      end_time: this.end_time.format(appConfig.datetime.format),
    };

    if (obj.type === "A/B") {
      // Interval should be null for A/B.
      obj.interval = null;
    }
    return obj;
  }

  /* stringify returns the Experiment's JSON string representation while handling
     big ints using jsonBig.stringify, that the default JSON.stringify fails to handle.
  */
  stringify() {
    return jsonBig.stringify(this.toJSON());
  }
}
