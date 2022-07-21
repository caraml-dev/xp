package util

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strconv"
)

const TypeCastingErrorTmpl = "invalid type of variable (%s) was provided for %s segmenter; expected %s"

// ConvertFloat64ToInt64 default interface type of float64 needs to be converted to expected int64 type
// required by proto
func ConvertFloat64ToInt64(value interface{}) (int64, bool) {
	val, ok := value.(float64)
	intValue := int64(val)

	return intValue, ok
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func DereferenceString(ref *string, default_ string) string {
	if ref == nil {
		return default_
	}

	return *ref
}

func DereferenceInt(ref *int64, default_ int64) int64 {
	if ref == nil {
		return default_
	}

	return *ref
}

func DereferenceUInt(ref *uint64, default_ uint64) uint64 {
	if ref == nil {
		return default_
	}

	return *ref
}

func DereferenceBool(ref *bool, default_ bool) bool {
	if ref == nil {
		return default_
	}

	return *ref
}

func GetFloatSegmenter(values map[string]interface{}, key string, segmenter string) (*float64, error) {
	var val float64
	var err error
	if reflect.TypeOf(values[key]).String() == "string" {
		val, err = strconv.ParseFloat(values[key].(string), 64)
		if err != nil {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, key, segmenter, "float64")
		}
	} else {
		castedVal, ok := values[key].(float64)
		if !ok {
			return nil, fmt.Errorf(TypeCastingErrorTmpl, key, segmenter, "float64")
		}
		val = castedVal
	}

	return &val, nil
}
