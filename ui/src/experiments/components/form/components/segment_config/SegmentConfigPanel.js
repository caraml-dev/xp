import React, { useContext, useEffect, useRef, useState } from "react";

import {
  EuiComboBox,
  EuiFlexGroup,
  EuiForm,
  EuiFormRow,
  EuiLoadingChart,
  EuiSpacer,
} from "@elastic/eui";
import { OverlayMask, get } from "@caraml-dev/ui-lib";

import { useXpApi } from "hooks/useXpApi";
import SegmenterContext from "providers/segmenter/context";

import { SegmenterConfigRow } from "./SegmenterConfigRow";

export const SegmentConfigPanel = ({
  projectId,
  segment,
  segmentTemplate,
  segmentSelectionOptions,
  onChange,
  errors = {},
}) => {
  const { dependencyMap, getSegmenterOptions, isLoading } =
    useContext(SegmenterContext);
  const options = getSegmenterOptions(segment);

  const [segmentId, setSegmentId] = useState();
  const [hasNewResponse, setHasNewResponse] = useState(false);

  // Set all the segmenter keys the first time
  useEffect(() => {
    const missingSegmenters = options.filter((e) => !(e.name in segment));
    if (missingSegmenters.length > 0) {
      onChange("segment")({
        ...missingSegmenters.reduce((acc, e) => {
          return { ...acc, [e.name]: [] };
        }, {}),
        ...segment,
      });
    }
  }, [options, segment, onChange]);

  // onChange handler for the individual segmenters also updates
  // dependent segmenters
  const onChangeSegmenterValue = (name, value) => {
    const dependentSegmenters = dependencyMap[name] || [];
    const depSegmentersSet = new Set(dependentSegmenters);
    // get a map of segmenter names as keys and the set of valid options as values
    const depOptions = options.reduce(
      (acc, cur) =>
        depSegmentersSet.has(cur.name)
          ? { ...acc, [cur.name]: new Set(cur.options.map((x) => x.value)) }
          : acc,
      {}
    );

    const updatedSegment = dependentSegmenters.reduce((acc, dep) => {
      // Reset dependent segmenter values by removing any invalid options
      return {
        ...acc,
        [dep]: segment[dep].filter((x) => depOptions[dep].has(x)),
      };
    }, segment);
    onChange("segment")({ ...updatedSegment, [name]: value });
  };

  // Ref for the overlay
  const segmentSectionRef = useRef();

  const [{ data: segmentDetails, isLoaded: isAPILoaded }, fetchSegmentDetails] =
    useXpApi(`/projects/${projectId}/segments/${segmentId}`, {}, {}, false);

  const onCustomOrTemplateSelection = (selected) => {
    const newSegment = options.reduce(function (map, obj) {
      map[obj.name] = [];
      return map;
    }, {});
    onChange("segment")(newSegment);

    // If id is not present, it will be set to undefined, but will not trigger useEffect
    setSegmentId(selected[0]?.id);
    onChange("segment_template")(selected.length > 0 ? selected[0] : "");
  };

  // Fetch Segment details every time there is a selection of Segment with id
  useEffect(() => {
    if (!!segmentId) {
      fetchSegmentDetails();
      setHasNewResponse(true);
    }
  }, [segmentId, fetchSegmentDetails]);

  // Populate Segment object so that it will reflect on UI
  useEffect(() => {
    if (hasNewResponse && isAPILoaded) {
      const activeSegmenters = options.map((opt) => opt.name);
      for (const [key, value] of Object.entries(segmentDetails.data.segment)) {
        // Update project's segmenters
        // Exclude inactive segmenters
        if (activeSegmenters.includes(key)) {
          onChange(`segment.${key}`)(value);
        }
      }
      setHasNewResponse(false);
    }
  }, [segmentDetails, options, onChange, hasNewResponse, isAPILoaded]);

  return !isLoading ? (
    <EuiForm>
      <EuiFormRow fullWidth label="Template">
        <EuiComboBox
          placeholder="Copy from Pre-configured Segment"
          isDisabled={
            !segmentSelectionOptions || segmentSelectionOptions.length === 0
          }
          fullWidth={true}
          singleSelection={{ asPlainText: true }}
          options={segmentSelectionOptions}
          onChange={onCustomOrTemplateSelection}
          selectedOptions={!!segmentTemplate ? [segmentTemplate] : []}
        />
      </EuiFormRow>
      <EuiSpacer />
      <EuiFlexGroup direction="column">
        {options.map((opt) => (
          <SegmenterConfigRow
            key={opt.name}
            name={opt.name}
            type={opt.type}
            description={opt.description}
            isRequired={opt.required}
            isMultiValued={opt.multi_valued}
            options={opt.options}
            values={segment[opt.name] || []}
            onChange={(value) => onChangeSegmenterValue(opt.name, value)}
            errors={get(errors.segment, opt.name)}
          />
        ))}
      </EuiFlexGroup>
    </EuiForm>
  ) : (
    <div ref={segmentSectionRef}>
      <OverlayMask parentRef={segmentSectionRef} opacity={0.4}>
        <EuiLoadingChart size="xl" mono />
      </OverlayMask>
    </div>
  );
};
