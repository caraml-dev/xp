package utils

import (
	"github.com/golang-collections/collections/set"
)

func IsUniqueStrings(strings []string) bool {
	values := make([]interface{}, len(strings))
	for i := range strings {
		values[i] = strings[i]
	}
	valueSet := set.New(values...)
	return valueSet.Len() == len(strings)
}

func StringSliceToSet(strings []string) *set.Set {
	stringSetInterface := make([]interface{}, len(strings))
	for i := range strings {
		stringSetInterface[i] = strings[i]
	}
	return set.New(stringSetInterface...)
}
