package util

import (
	"hash/fnv"
)

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
