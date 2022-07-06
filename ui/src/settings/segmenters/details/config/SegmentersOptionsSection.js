import { EuiPanel } from "@elastic/eui";

import { CodeBlockPanel } from "settings/segmenters/details/config/component/CodeBlockPanel";
import { convertObjectToString } from "utils/helpers";

export const SegmentersOptionsSection = ({ options }) => {
  return (
    <EuiPanel>
      <CodeBlockPanel text={convertObjectToString(options)} />
    </EuiPanel>
  );
};
