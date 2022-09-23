import React, { useEffect, useState } from "react";

import intersection from "lodash/intersection";
import isEmpty from "lodash/isEmpty";
import union from "lodash/union";

import { useXpApi } from "hooks/useXpApi";

const SegmenterContext = React.createContext({});

export const SegmenterContextProvider = ({ projectId, children, status }) => {
  const [
    {
      data: { data: segmenters },
      isLoading,
      isLoaded,
    },
  ] = useXpApi(
    `/projects/${projectId}/segmenters`,
    {
      query: {
        status: status,
      },
    },
    []
  );

  // Map of segmenter name -> [dependent segmenter names]
  const [dependencyMap, setDependencyMap] = useState({});
  // Contains the information about the segmenter
  const [segmenterConfig, setSegmenterConfig] = useState([]);
  // isSegmenterConfigLoaded indicates if all the internal processing has been completed
  const [isSegmenterConfigLoaded, setIsSegmenterConfigLoaded] = useState(false);

  useEffect(() => {
    // Clear isSegmenterConfigLoaded if isLoaded is cleared
    if (!isLoaded && isSegmenterConfigLoaded) {
      setIsSegmenterConfigLoaded(false);
    }
  }, [isLoaded, isSegmenterConfigLoaded]);

  useEffect(() => {
    if (isLoaded) {
      let deps = {};
      let segmentersName = [];
      segmenters.forEach((s) => {
        segmentersName.push({
          name: s.name,
          required: s.required,
          variables: s.treatment_request_fields,
          status: s.status,
          scope: s.scope,
        });
        if (!!s.constraints) {
          s.constraints.forEach((c) => {
            c.pre_requisites.forEach((p) => {
              deps[p.segmenter_name] = Array.from(
                new Set([...(deps[p.segmenter_name] || []), s.name])
              );
            });
          });
        }
      });
      setDependencyMap(deps);
      setSegmenterConfig(segmentersName);
      setIsSegmenterConfigLoaded(true);
    }
  }, [segmenters, isLoaded]);

  // getSegmenterOptions returns the segmenter info for the project,
  // where the options are filtered based on any currently chosen segmenter values.
  const getSegmenterOptions = (currentValues) => {
    if (isLoaded) {
      return segmenters.map((e) => {
        let allowedValues = [];
        let constrainedOptions = {};
        // If constraints set, evaluate them against currentValues and gather the
        // allowed values.
        if (!!e.constraints) {
          e.constraints.forEach((c) => {
            c.pre_requisites.forEach((p) => {
              if (
                intersection(
                  currentValues[p.segmenter_name],
                  p.segmenter_values
                ).length > 0
              ) {
                allowedValues = union(allowedValues, c.allowed_values);
                if (!!c.options) {
                  // Add to constraint-specific options
                  Object.entries(c.options).forEach(([k, v]) => {
                    constrainedOptions[k] = v;
                  });
                }
              }
            });
          });
        }
        // allowedValues can either be empty (in which case we allow all values)
        // or the options will have to be limited to the ones in allowedValues
        const options = isEmpty(constrainedOptions)
          ? e.options
          : constrainedOptions;
        const displayOptions = Object.keys(options).reduce((acc, key) => {
          if (!allowedValues.length || allowedValues.includes(options[key])) {
            return [
              ...acc,
              {
                value: options[key],
                label: key,
              },
            ];
          }
          return acc;
        }, []);
        // Replace raw options with value and label
        return { ...e, options: displayOptions };
      });
    }
    return [];
  };

  return (
    <SegmenterContext.Provider
      value={{
        getSegmenterOptions,
        dependencyMap,
        segmenterConfig,
        isLoading,
        isLoaded: isSegmenterConfigLoaded,
      }}>
      {children}
    </SegmenterContext.Provider>
  );
};

export default SegmenterContext;
