import React, { useMemo, useRef } from "react";

import {
  EuiCallOut,
  EuiDescriptionList,
  EuiDescriptionListDescription,
  EuiDescriptionListTitle,
  EuiLink,
  EuiLoadingChart,
  EuiSpacer,
  EuiTextColor,
  EuiTitle,
} from "@elastic/eui";
import { OverlayMask } from "@caraml-dev/ui-lib";
import moment from "moment";

import { useConfig } from "config";
import { useXpApi } from "hooks/useXpApi";

import "./ExperimentsConfigGroup.scss";

export const ExperimentsConfigGroup = ({ projectId }) => {
  const { appConfig } = useConfig();
  const timestamp = useMemo(
    () => moment.utc().format(appConfig.datetime.format),
    [appConfig]
  );
  // Get experiments that are currently active
  const [{ data, isLoading, error }] = useXpApi(
    `/projects/${projectId}/experiments`,
    {
      query: {
        start_time: timestamp,
        end_time: timestamp,
        status: "active",
      },
    },
    { data: [], paging: { total: 0 } }
  );

  const summaryOverlayRef = useRef();

  /* The below block uses `EuiTextColor` without any special color,
     because there are some subtle style differences with `EuiText`. */
  return (
    <EuiTitle size="xs">
      <EuiTextColor>
        <EuiSpacer size="s" />
        <EuiLink
          href={`/turing/projects/${projectId}/experiments`}
          target="_blank"
          external>
          Experiments
        </EuiLink>
        <EuiSpacer size="s" />
        {!!error ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert">
            <p>{error.message}</p>
          </EuiCallOut>
        ) : isLoading ? (
          <div ref={summaryOverlayRef}>
            <OverlayMask parentRef={summaryOverlayRef} opacity={0.4}>
              <EuiLoadingChart size="l" mono />
            </OverlayMask>
          </div>
        ) : (
          <EuiDescriptionList
            textStyle="reverse"
            type="responsiveColumn"
            compressed>
            <EuiDescriptionListTitle>
              Active Experiments
            </EuiDescriptionListTitle>
            <EuiDescriptionListDescription>
              {data.paging.total}
            </EuiDescriptionListDescription>
          </EuiDescriptionList>
        )}
      </EuiTextColor>
    </EuiTitle>
  );
};
