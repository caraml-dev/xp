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
