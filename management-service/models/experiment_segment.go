package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/caraml-dev/xp/common/api/schema"
	_segmenters "github.com/caraml-dev/xp/common/segmenters"
	_utils "github.com/caraml-dev/xp/common/utils"
)

type ExperimentSegmentRaw map[string]interface{}
type ExperimentSegment map[string][]string

func (s *ExperimentSegment) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &s)
}

func (s ExperimentSegment) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// ToApiSchema converts all DB string values to appropriate ExperimentSegment values based on
// registered SegmenterType to be used when returning API response
func (s ExperimentSegment) ToApiSchema(segmentersType map[string]schema.SegmenterType) schema.ExperimentSegment {
	experimentSegment := schema.ExperimentSegment{}
	for key, vals := range s {
		switch segmentersType[key] {
		case schema.SegmenterTypeString:
			experimentSegment[key] = vals
		case schema.SegmenterTypeInteger:
			intVals := []int64{}
			for _, val := range vals {
				intVal, _ := strconv.Atoi(val)
				intVals = append(intVals, int64(intVal))
			}
			experimentSegment[key] = intVals
		case schema.SegmenterTypeReal:
			floatVals := []float64{}
			for _, val := range vals {
				float64Val, _ := strconv.ParseFloat(val, 64)
				floatVals = append(floatVals, float64Val)
			}
			experimentSegment[key] = floatVals
		case schema.SegmenterTypeBool:
			boolVals := []bool{}
			for _, val := range vals {
				boolVal, _ := strconv.ParseBool(val)
				boolVals = append(boolVals, boolVal)
			}
			experimentSegment[key] = boolVals
		}
	}

	return experimentSegment
}

// ToProtoSchema converts all DB string values to appropriate ListSegmenterValue based on
// registered SegmenterType to be used when sending messages to Treatment Service
func (s ExperimentSegment) ToProtoSchema(segmenterTypes map[string]schema.SegmenterType) map[string]*_segmenters.ListSegmenterValue {
	protoSegments := make(map[string]*_segmenters.ListSegmenterValue)
	for key, vals := range s {
		if len(vals) > 0 {
			switch segmenterTypes[key] {
			case schema.SegmenterTypeString:
				protoSegments[key] = _utils.StringSliceToListSegmenterValue(&vals)
			case schema.SegmenterTypeInteger:
				intVals := []int64{}
				for _, val := range vals {
					intVal, _ := strconv.Atoi(val)
					intVals = append(intVals, int64(intVal))
				}
				protoSegments[key] = _utils.Int64ListToListSegmenterValue(&intVals)
			case schema.SegmenterTypeReal:
				floatVals := []float64{}
				for _, val := range vals {
					float64Val, _ := strconv.ParseFloat(val, 64)
					floatVals = append(floatVals, float64Val)
				}
				protoSegments[key] = _utils.FloatListToListSegmenterValue(&floatVals)
			case schema.SegmenterTypeBool:
				boolVals := []bool{}
				for _, val := range vals {
					boolVal, _ := strconv.ParseBool(val)
					boolVals = append(boolVals, boolVal)
				}
				protoSegments[key] = _utils.BoolSliceToListSegmenterValue(&boolVals)
			}
		}
	}

	return protoSegments
}

// ToStorageSchema converts raw request ExperimentSegment values to string values for storing in DB
func (s ExperimentSegmentRaw) ToStorageSchema(segmenterTypes map[string]schema.SegmenterType) (ExperimentSegment, error) {
	segmenterVals := ExperimentSegment{}
	for k, v := range s {
		errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", k)
		vals := v.([]interface{})
		switch segmenterTypes[k] {
		case schema.SegmenterTypeString:
			strVals := []string{}
			for _, val := range vals {
				stringVal, ok := val.(string)
				if !ok {
					return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeString)
				}
				strVals = append(strVals, stringVal)
			}
			segmenterVals[k] = strVals
		case schema.SegmenterTypeInteger:
			strVals := []string{}
			for _, val := range vals {
				floatVal, ok := val.(float64)
				if !ok {
					return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
				}
				strVals = append(strVals, strconv.Itoa(int(floatVal)))
			}
			segmenterVals[k] = strVals
		case schema.SegmenterTypeReal:
			strVals := []string{}
			for _, val := range vals {
				floatVal, ok := val.(float64)
				if !ok {
					return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
				}
				strVals = append(strVals, strconv.FormatFloat(floatVal, 'f', -1, 64))
			}
			segmenterVals[k] = strVals
		case schema.SegmenterTypeBool:
			strVals := []string{}
			for _, val := range vals {
				boolValue, ok := val.(bool)
				if !ok {
					return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
				}
				strVals = append(strVals, strconv.FormatBool(boolValue))
			}
			segmenterVals[k] = strVals
		}
	}
	return segmenterVals, nil
}

// ToRawSchema converts ExperimentSegment string values from the DB into their actual value types. Optional segmenters
// will automatically be removed by this method.
func (s ExperimentSegment) ToRawSchema(segmentersType map[string]schema.SegmenterType) (ExperimentSegmentRaw, error) {
	rawSegments := ExperimentSegmentRaw{}
	for key, vals := range s {
		errTmpl := fmt.Sprintf("received wrong type of segmenter value; %s expects type", key)
		if len(vals) > 0 {
			switch segmentersType[key] {
			case schema.SegmenterTypeString:
				stringVals := []interface{}{}
				for _, val := range vals {
					stringVals = append(stringVals, val)
				}
				rawSegments[key] = stringVals
			case schema.SegmenterTypeInteger:
				// Raw Schema refers to JSON and numbers are treated as float64
				floatVals := []interface{}{}
				for _, val := range vals {
					_, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeInteger)
					}
					float64Val, _ := strconv.ParseFloat(val, 64)
					floatVals = append(floatVals, float64Val)
				}
				rawSegments[key] = floatVals
			case schema.SegmenterTypeReal:
				floatVals := []interface{}{}
				for _, val := range vals {
					float64Val, err := strconv.ParseFloat(val, 64)
					if err != nil {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeReal)
					}
					floatVals = append(floatVals, float64Val)
				}
				rawSegments[key] = floatVals
			case schema.SegmenterTypeBool:
				boolVals := []interface{}{}
				for _, val := range vals {
					boolVal, err := strconv.ParseBool(val)
					if err != nil {
						return nil, fmt.Errorf("%s %s", errTmpl, schema.SegmenterTypeBool)
					}
					boolVals = append(boolVals, boolVal)
				}
				rawSegments[key] = boolVals
			}
		}
	}

	return rawSegments, nil
}
