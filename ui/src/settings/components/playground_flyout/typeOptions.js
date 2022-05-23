export const validation_url = "validation_url";
export const treatment_schema = "treatment_schema";

export const getValidationOptions = (settings) => [
  {
    id: validation_url,
    label: "External Validation",
    disabled: !!settings && settings.validation_url === "",
    placeholderText: `Enter your sample request payload. Eg: \n${JSON.stringify(
      {
        entity_type: "treatment",
        operation: "create",
        data: {
          field1: "abc",
          field2: "def",
          field3: {
            field4: 0.1,
          },
        },
      },
      null,
      4
    )}`,
  },
  {
    id: treatment_schema,
    label: "Treatment Validation Rules",
    disabled: !!settings && settings.treatment_schema.rules.length === 0,
    placeholderText: `Enter your sample treatment configuration. Eg: \n${JSON.stringify(
      {
        field1: "abc",
        field2: "Def",
        field3: {
          field4: 0.1,
        },
      },
      null,
      4
    )}`,
  },
];
