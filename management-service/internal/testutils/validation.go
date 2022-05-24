package testutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

// AssertEqualValues helps compare values recursively, using the Equal method where exists.
// This helps with timestamp comparisons.
func AssertEqualValues(t *testing.T, wanted, got interface{}) {
	equality := cmp.Equal(wanted, got)
	assert.True(t, equality)

	if !equality {
		// Log diff
		t.Errorf(cmp.Diff(wanted, got))
	}
}
