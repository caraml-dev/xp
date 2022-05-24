package util

import (
	"errors"

	"github.com/golang/geo/s2"
)

func GetS2ID(lat float64, long float64, level int) (s2.CellID, error) {
	latLng := s2.LatLngFromDegrees(lat, long)
	if !latLng.IsValid() {
		return s2.CellID(1), errors.New("received invalid latitude, longitude values")
	}

	if !isValidLevel(level) {
		return s2.CellID(1), errors.New("received invalid s2 geo level")
	}

	cell := s2.CellIDFromLatLng(latLng).Parent(level)
	return cell, nil
}

func isValidLevel(level int) bool {
	if level <= 30 && level > -1 {
		return true
	}
	return false
}
