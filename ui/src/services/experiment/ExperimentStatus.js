import moment from "moment";

import { appConfig } from "config";

export const getExperimentStatus = (experiment) => {
  /*
      API provide status as inactive or active. For Active there is 3 possible states.
      Deactivated - status is inactive
      Completed - end time is before current date
      Schedule - start time after current date
      Active - current date in between start and end time
    */
  const statusMapping = {
    Deactivated: {
      label: "Deactivated",
      color: "default",
      iconType: "cross",
    },
    Completed: {
      label: "Completed",
      color: "success",
      iconType: "check",
    },
    Scheduled: {
      label: "Scheduled",
      color: "warning",
      iconType: "calendar",
    },
    Running: {
      label: "Running",
      color: "primary",
      iconType: "clock",
    },
  };

  if (experiment.status === "inactive") {
    return statusMapping["Deactivated"];
  }

  var startTime = moment(experiment.start_time, appConfig.datetime.format);
  var endTime = moment(experiment.end_time, appConfig.datetime.format);

  if (endTime.isBefore()) {
    return statusMapping["Completed"];
  }
  if (startTime.isAfter()) {
    return statusMapping["Scheduled"];
  }
  return statusMapping["Running"];
};
