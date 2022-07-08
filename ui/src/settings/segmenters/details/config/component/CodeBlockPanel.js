import { Fragment } from "react";

import { EuiCodeBlock, EuiPanel } from "@elastic/eui";

export const CodeBlockPanel = ({ text }) => {
  // Code block cant be used for empty text as expansion icon will still be present
  // default text to "-" if empty
  return (
    <Fragment>
      {text !== "" ? (
        <EuiPanel hasBorder={false} color="subdued" style={{ height: "100%" }}>
          <EuiCodeBlock
            className="eui-textBreakWord"
            isCopyable={true}
            overflowHeight={200}
            transparentBackground={true}
            fontSize="m">
            {text}
          </EuiCodeBlock>
        </EuiPanel>
      ) : (
        "-"
      )}
    </Fragment>
  );
};
