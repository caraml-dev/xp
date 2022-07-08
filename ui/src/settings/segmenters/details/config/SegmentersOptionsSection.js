import { EuiPanel } from "@elastic/eui";

import { convertObjectToString } from "utils/helpers";

import { CodeBlockPanel } from "./component/CodeBlockPanel";

export const SegmentersOptionsSection = ({ options }) => {
  return (
    <EuiPanel>
      <CodeBlockPanel text={convertObjectToString(options)} />
    </EuiPanel>
  );
};
