import { Fragment, useState } from "react";

import {
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiPanel,
  EuiSpacer,
  EuiText,
  EuiToolTip,
} from "@elastic/eui";

import { CodeBlockPanel } from "settings/segmenters/details/config/component/CodeBlockPanel";
import {
  convertArrayToString,
  convertObjectToString,
  formatJsonString,
} from "utils/helpers";

export const SegmentersConstraintSection = ({ constraints }) => {
  const formattedConstraints = constraints.map((constraint) => {
    return {
      allowed_values: convertArrayToString(constraint.allowed_values),
      pre_requisites: formatJsonString(constraint.pre_requisites),
      options: convertObjectToString(constraint.options),
    };
  });

  return (
    <EuiPanel>
      {constraints.length > 0
        ? formattedConstraints.map((constraint, idx) => (
            <SegmenterConstraintPanel
              key={idx}
              constraint={constraint}></SegmenterConstraintPanel>
          ))
        : "-"}
    </EuiPanel>
  );
};

const SegmenterConstraintPanel = ({ constraint }) => {
  const [toggle, setToggle] = useState(false);
  // when options are available, buttonIcon is shown and allowedValues section does not need padding
  // while options have to be padded
  const allowedValuesPaddingStyle =
    constraint.options !== "" ? {} : { paddingRight: "45px" };
  const optionsPaddingStyle =
    constraint.options !== "" ? { paddingRight: "45px" } : {};
  return (
    <Fragment>
      <EuiPanel>
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <EuiText size="xs" style={{ fontWeight: "bold" }}>
              Pre Requisites
            </EuiText>
            <EuiSpacer size="xs"></EuiSpacer>
            <CodeBlockPanel text={constraint.pre_requisites} />
          </EuiFlexItem>
          <EuiFlexItem grow={1} style={allowedValuesPaddingStyle}>
            <EuiText size="xs" style={{ fontWeight: "bold" }}>
              Allowed Values
            </EuiText>
            <EuiSpacer size="xs"></EuiSpacer>
            <CodeBlockPanel text={constraint.allowed_values} />
          </EuiFlexItem>
          {constraint.options !== "" && (
            <EuiFlexItem grow={false}>
              <EuiToolTip content="Value Overrides">
                <EuiButtonIcon
                  iconType="indexSettings"
                  aria-label="Value override"
                  onClick={() => {
                    setToggle(!toggle);
                  }}
                />
              </EuiToolTip>
            </EuiFlexItem>
          )}
        </EuiFlexGroup>
        {toggle && (
          <EuiFlexGroup>
            <EuiFlexItem style={optionsPaddingStyle}>
              <EuiSpacer size="xs"></EuiSpacer>
              <EuiText size="xs" style={{ fontWeight: "bold" }}>
                Value Overrides
              </EuiText>
              <CodeBlockPanel text={constraint.options} />
            </EuiFlexItem>
          </EuiFlexGroup>
        )}
        <EuiSpacer />
      </EuiPanel>
      <EuiSpacer />
    </Fragment>
  );
};
