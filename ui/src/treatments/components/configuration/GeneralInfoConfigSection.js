import { React } from "react";

import { EuiCodeBlock, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";
import { formatJsonString } from "utils/helpers";

export const GeneralInfoConfigSection = ({ treatment }) => {
  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem>
        <ConfigPanel>
          <EuiCodeBlock
            language="json"
            fontSize="s"
            paddingSize="s"
            overflowHeight={500}
            style={{ minHeight: 50 }}
            isCopyable>
            {formatJsonString(treatment.configuration)}
          </EuiCodeBlock>
        </ConfigPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
