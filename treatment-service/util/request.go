package util

import (
	"fmt"
)

func ValidateFloat64Type(value interface{}, name string) error {
	switch value.(type) {
	case float64:
		break
	default:
		return fmt.Errorf("incorrect type provided for %s; expected float64", name)
	}
	return nil
}

func ValidateLatLong(filterParams map[string]interface{}) error {
	if err := ValidateFloat64Type(filterParams["latitude"], "latitude"); err != nil {
		return err
	}
	if err := ValidateFloat64Type(filterParams["longitude"], "longitude"); err != nil {
		return err
	}
	return nil
}
