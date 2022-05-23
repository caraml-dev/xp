import { React } from "react";

import { EuiCodeBlock, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";

import { ConfigPanel } from "components/config_section/ConfigPanel";

export const GeneralInfoConfigSection = ({ treatment }) => {
  const config = JSON.stringify(treatment.configuration);
  const formattedText = JSON.stringify(JSON.parse(config), null, 2);
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
            {formattedText}
          </EuiCodeBlock>
        </ConfigPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
