package segmenters

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	_segmenters "github.com/caraml-dev/xp/common/segmenters"
)

type Runner interface {
	GetName() string
	Transform(name string, requestValues map[string]interface{}, experimentVariables []string) ([]*_segmenters.SegmenterValue, error)
}

var runnersLock sync.Mutex

// Runners contain all the registered segmenter runners by name.
var Runners = make(map[string]Factory)

// Factory creates a segmenter runner from the provided config.
//
// Config is a raw encoded JSON value. The segmenter runner implementation
// for each segmenter should provide a schema and example
// of the JSON value to explain the usage.
type Factory func(config json.RawMessage) (Runner, error)

// Register a segmenter runner with the provided name and factory function.
//
// For registration to be properly recorded, Register function should be called in the init
// phase of the Go execution. The init function is usually defined in the package where
// the segmenter runner is implemented. The name of the segmenter runners should be unique
// across all implementations. Registering multiple segmenter runners with the
// same name will return an error.
func Register(name string, factory Factory) error {
	runnersLock.Lock()
	defer runnersLock.Unlock()

	name = strings.ToLower(name)
	if _, found := Runners[name]; found {
		return fmt.Errorf("segmenter runner %q was registered twice", name)
	}

	Runners[name] = factory
	return nil
}

// Get a segmenter runner that has been registered.
//
// The segmenter runner will be initialized using the registered factory function with the provided config.
// Retrieving a segmenter runner that is not yet registered will return an error.
func Get(name string, config json.RawMessage) (Runner, error) {
	runnersLock.Lock()
	defer runnersLock.Unlock()

	name = strings.ToLower(name)
	m, ok := Runners[name]
	if !ok {
		return nil, fmt.Errorf("no segmenter runner found for name %s", name)
	}

	return m(config)
}
