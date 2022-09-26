import { useContext, useMemo, useRef, useState } from "react";

import {
  EuiFlexGroup,
  EuiInMemoryTable,
  EuiLoadingChart,
  EuiText,
  EuiTextAlign,
} from "@elastic/eui";
import { useDimension, useToggle } from "@gojek/mlp-ui";

import { ConfigPanel } from "components/config_section/ConfigPanel";
import { ConfigSectionFlyout } from "components/config_section/ConfigSectionFlyout";
import { ExpandableTableColumn } from "components/table/ExpandableTableColumn";
import SegmenterContext from "providers/segmenter/context";
import { stringifySegmenterValue } from "services/experiment/Segment";

export const SegmentConfigSection = ({ experiment }) => {
  const [isFlyoutVisible, toggleFlyout] = useToggle();
  const [flyoutItem, setFlyoutItem] = useState();
  const openFlyout = (item) => () => {
    const formattedItem = {
      segmentName: item.segmentName,
      segmentValue: item.segmentValues.join(",\r\n"),
    };
    setFlyoutItem(formattedItem);
    toggleFlyout();
  };

  //Fetch Segmenter options from context which calls get project segmenters API
  const { isLoaded, getSegmenterOptions, segmenterConfig } =
    useContext(SegmenterContext);
  const segmenterOptions = getSegmenterOptions(experiment.segment);

  //Create a map with segment name as key, nested map with {segmentValues:label} as value
  const segmenterOptionsMapping = segmenterOptions.reduce((acc, entry) => {
    acc[entry.name] = entry.options.reduce((dict, option) => {
      dict[option.value] = option.label;
      return dict;
    }, {});
    return acc;
  }, {});

  //Take in segment name as key, array as values to be mapped into labels
  const formatSegmenterValue = (segmenterName, arrValue) => {
    return segmenterName in segmenterOptionsMapping
      ? arrValue.map((id) => segmenterOptionsMapping[segmenterName][id] || id)
      : arrValue;
  };

  //Create a formatted map of segmenters with name as key and array values as value
  const items = Object.entries(experiment.segment).map(
    ([segmentName, segmentValues]) => ({
      segmentName,
      segmentValues: formatSegmenterValue(segmentName, segmentValues),
    })
  );

  const activeProjectSegmenters = useMemo(() => {
    return segmenterConfig.reduce((acc, segmenter) => {
      if (segmenter.status === "active") {
        acc.push(segmenter.name);
      }
      return acc;
    }, []);
  }, [segmenterConfig]);

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : (
    <ConfigPanel>
      <ExperimentSegmentTable
        items={items}
        projectSegmenters={activeProjectSegmenters}
        buttonAction={openFlyout}
      />
      {isFlyoutVisible && (
        <ConfigSectionFlyout
          header={flyoutItem.segmentName}
          content={flyoutItem.segmentValue}
          onClose={toggleFlyout}
          size="s"
          contentClass="eui-textBreakWord"
          textStyle={{ textTransform: "capitalize" }}
        />
      )}
    </ConfigPanel>
  );
};

const ExperimentSegmentTable = ({ items, projectSegmenters, buttonAction }) => {
  const segmentValueColumnRef = useRef();
  const { width: contentWidth } = useDimension(segmentValueColumnRef);

  const columns = [
    {
      field: "segmentName",
      width: "20%",
      name: "Name",
      render: (segmentName, item) => {
        return (
          <span className="eui-textTruncate">
            <EuiText
              className="eui-textTruncate"
              size="s"
              // Colors need to be set in style, or it will overwrite the truncate properties
              style={{
                color: projectSegmenters.includes(item.segmentName)
                  ? "#1a1c21"
                  : "#bd271e",
              }}>
              {segmentName}
            </EuiText>
          </span>
        );
      },
    },
    {
      field: "segmentValues",
      width: "80%",
      name: "Value",
      render: (segmentValues, item) => {
        return (
          <EuiFlexGroup
            ref={segmentValueColumnRef}
            className="eui-textTruncate">
            <ExpandableTableColumn
              text={
                segmentValues
                  .map((v) => stringifySegmenterValue(v))
                  .join(", ") || "-"
              }
              buttonAction={buttonAction(item)}
              allowedWidth={contentWidth}
              textStyle={{
                fontWeight: "bold",
                textTransform: "capitalize",
                color: projectSegmenters.includes(item.segmentName)
                  ? "#1a1c21"
                  : "#bd271e",
              }}
            />
          </EuiFlexGroup>
        );
      },
    },
  ];

  return <EuiInMemoryTable items={items} columns={columns} />;
};
