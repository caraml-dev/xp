import { EuiPanel } from "@elastic/eui";

import { CodeBlockPanel } from "settings/segmenters/details/config/component/CodeBlockPanel";
import { convertArrayToString } from "utils/helpers";

export const SegmentersTreatmentReqSection = ({ treatmentRequestFields }) => {
  return (
    <EuiPanel>
      <CodeBlockPanel text={convertArrayToString(treatmentRequestFields)} />
    </EuiPanel>
  );
};
