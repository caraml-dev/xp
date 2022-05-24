import React, { Fragment } from "react";

import { EuiSuperSelect, EuiText } from "@elastic/eui";

const SuperSelectWithDescription = ({ value, onChange, options, ...props }) => {
  const optionsWithDescription = options.map((e) => {
    let option = { ...e, inputDisplay: e.inputDisplay || e.label };
    if (!!option.description) {
      option.dropdownDisplay = (
        <Fragment>
          <strong>{option.inputDisplay}</strong>
          <EuiText size="s" color="subdued">
            <p className="euiTextColor--subdued">{e.description}</p>
          </EuiText>
        </Fragment>
      );
    }
    return option;
  });

  return (
    <EuiSuperSelect
      disabled={props.disabled}
      options={optionsWithDescription}
      valueOfSelected={value}
      onChange={onChange}
      itemLayoutAlign={props.itemLayoutAlign || "top"}
      hasDividers={!!props.hasDividers}
      isInvalid={props.isInvalid}
      fullWidth={props.fullWidth}
      compressed={props.compressed}
    />
  );
};

export default SuperSelectWithDescription;
