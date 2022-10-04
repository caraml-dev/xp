export const experimentTypes = [
  {
    value: "A/B",
    label: "A/B",
    description: "Treatments are randomized on the Randomization Unit",
  },
  {
    value: "Switchback",
    label: "Switchback",
    description: "Treatments are determined by the Switchback Interval",
  },
];
export const experimentStatuses = [
  {
    value: "active",
    label: "Active",
  },
  {
    value: "inactive",
    label: "Inactive",
  },
];

export const experimentStatusesFriendly = [
  {
    value: "scheduled",
    label: "Scheduled",
    color: "warning",
    iconType: "calendar",

  },
  {
    value: "running",
    label: "Running",
    color: "primary",
    iconType: "clock",
  },
  {
    value: "completed",
    label: "Completed",
    color: "success",
    iconType: "check",
  },
  {
    value: "deactivated",
    label: "Deactivated",
    color: "default",
    iconType: "cross",
  },
];

export const experimentTiers = [
  {
    value: "default",
    label: "Default",
  },
  {
    value: "override",
    label: "Override",
  },
];
