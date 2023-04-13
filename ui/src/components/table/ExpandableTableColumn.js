import { useRef } from "react";

import {
  EuiButtonIcon,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
} from "@elastic/eui";
import { useDimension } from "@caraml-dev/ui-lib";

export const ExpandableTableColumn = ({
  buttonAction,
  text,
  textStyle,
  allowedWidth,
}) => {
  const columnRef = useRef();
  const { width: contentWidth } = useDimension(columnRef);

  // allowedWidth is rounded down, as there could be very small precision off the calculation of columnRef causing the view to toggle infinitely
  return contentWidth < Math.floor(allowedWidth) ? (
    // Span is used instead to measure actual size, if overflow, switch to truncate style with button
    <span ref={columnRef}>
      <EuiFlexItem grow={true}>
        <EuiText
          size="s"
          style={
            textStyle
              ? textStyle
              : {
                  fontWeight: "bold",
                }
          }>
          {text}
        </EuiText>
      </EuiFlexItem>
    </span>
  ) : (
    <span ref={columnRef} className="eui-textTruncate">
      <EuiFlexGroup direction="row">
        <EuiFlexItem className="eui-textTruncate">
          <EuiText
            className="eui-textTruncate"
            size="s"
            style={
              textStyle
                ? textStyle
                : {
                    fontWeight: "bold",
                  }
            }>
            {text}
          </EuiText>
        </EuiFlexItem>
        <EuiFlexItem grow={false}>
          <EuiButtonIcon
            iconType="arrowRight"
            onClick={buttonAction}
            aria-label="Open Flyout"
          />
        </EuiFlexItem>
      </EuiFlexGroup>
    </span>
  );
};
