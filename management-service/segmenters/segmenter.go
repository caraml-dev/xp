package segmenters

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type Segmenter interface {
	GetName() string
	GetType() _segmenters.SegmenterValueType
	GetConfiguration() (*_segmenters.SegmenterConfiguration, error)
	GetExperimentVariables() *_segmenters.ListExperimentVariables
	IsValidType(inputValues []*_segmenters.SegmenterValue) bool
	ValidateSegmenterAndConstraints(segment map[string]*_segmenters.ListSegmenterValue) error
}

var segmentersLock sync.Mutex

// segmenters contain all the registered experiment segmenters by name.
var Segmenters = make(map[string]Factory)

// Factory creates a segmenter manager from the provided config.
//
// Config is a raw encoded JSON value. The segmenter manager implementation
// for each segmenter should provide a schema and example
// of the JSON value to explain the usage.
type Factory func(config json.RawMessage) (Segmenter, error)

// Register an experiment segmenter with the provided name and factory function.
//
// For registration to be properly recorded, Register function should be called in the init
// phase of the Go execution. The init function is usually defined in the package where
// the segmenter is implemented. The name of the experiment segmenters should be unique
// across all implementations. Registering multiple experiment segmenters with the
// same name will return an error.
func Register(name string, factory Factory) error {
	segmentersLock.Lock()
	defer segmentersLock.Unlock()

	name = strings.ToLower(name)
	if _, found := Segmenters[name]; found {
		return fmt.Errorf("segmenter %q was registered twice", name)
	}

	Segmenters[name] = factory
	return nil
}

// Get an experiment segmenter that has been registered.
//
// The segmenter will be initialized using the registered factory function with the provided config.
// Retrieving an experiment segmenter that is not yet registered will return an error.
func Get(name string, config json.RawMessage) (Segmenter, error) {
	segmentersLock.Lock()
	defer segmentersLock.Unlock()

	name = strings.ToLower(name)
	m, ok := Segmenters[name]
	if !ok {
		return nil, fmt.Errorf("no segmenter found for name %s", name)
	}

	return m(config)
}
