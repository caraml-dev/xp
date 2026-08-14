package models

import (
	"testing"
	"time"

	"github.com/caraml-dev/mlp/api/pkg/instrumentation/metrics"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
	"github.com/caraml-dev/xp/treatment-service/instrumentation"
)

// fakeMetricsRecorder is a test double for LocalStorageMetricsRecorder (i.e. for
// services.MetricService's LogRequestCount/LogLatencyHistogram), recording every call it
// receives so tests can assert on ordering and arguments without depending on any real
// metrics backend (Prometheus, RPC, ...).
type fakeMetricsRecorder struct {
	callCounts     map[string]int
	durationCounts map[string]int
}

func newFakeMetricsRecorder() *fakeMetricsRecorder {
	return &fakeMetricsRecorder{
		callCounts:     map[string]int{},
		durationCounts: map[string]int{},
	}
}

func (f *fakeMetricsRecorder) LogRequestCount(labels map[string]string, loggingMetric metrics.MetricName) {
	if loggingMetric != instrumentation.LocalStorageCallCount {
		return
	}
	f.callCounts[labels["method"]]++
}

func (f *fakeMetricsRecorder) LogLatencyHistogram(begin time.Time, labels map[string]string, loggingMetric metrics.MetricName) {
	if loggingMetric != instrumentation.LocalStorageCallDurationMs {
		return
	}
	f.durationCounts[labels["method"]]++
}

// TestInstrumentCall verifies the shared helper's contract directly: the counter reports
// immediately when instrumentCall is invoked, and the duration only reports once the returned
// func is actually called (mirroring how it's used: called immediately, deferred return value).
func TestInstrumentCall(t *testing.T) {
	recorder := newFakeMetricsRecorder()
	s := &LocalStorage{}
	s.SetMetricsRecorder(recorder)

	done := s.instrumentCall("TestMethod")
	require.Equal(t, 1, recorder.callCounts["TestMethod"], "call count should report immediately")
	require.Equal(t, 0, recorder.durationCounts["TestMethod"], "duration should not report until done() runs")

	done()
	require.Equal(t, 1, recorder.durationCounts["TestMethod"], "duration should report once done() runs")
}

// TestInstrumentCall_NoRecorder verifies that with no recorder set (the default), instrumentCall
// returns a harmless no-op -- this is what every public method relies on to skip metrics
// entirely when nothing has called SetMetricsRecorder.
func TestInstrumentCall_NoRecorder(t *testing.T) {
	s := &LocalStorage{}
	done := s.instrumentCall("TestMethod")
	require.NotPanics(t, done)
}

// TestFindExperiments_MetricsGatedByRecorder exercises a real read-path public method end to
// end: with no recorder set (matching every existing caller/test), nothing is reported; once
// SetMetricsRecorder is called, the call is counted and its duration reported.
func TestFindExperiments_MetricsGatedByRecorder(t *testing.T) {
	method := "FindExperiments"

	t.Run("no recorder by default", func(t *testing.T) {
		s := &LocalStorage{Experiments: map[ProjectId][]*ExperimentIndex{}}
		s.FindExperiments(0, nil)
		// No recorder set -- nothing to assert beyond "this didn't panic".
	})

	t.Run("recorder set", func(t *testing.T) {
		recorder := newFakeMetricsRecorder()
		s := &LocalStorage{Experiments: map[ProjectId][]*ExperimentIndex{}}
		s.SetMetricsRecorder(recorder)

		s.FindExperiments(0, nil)

		require.Equal(t, 1, recorder.callCounts[method])
		require.Equal(t, 1, recorder.durationCounts[method])
	})
}

// TestInsertExperiment_MetricsGatedByRecorder mirrors the above for a write-path public
// method, to confirm the same pattern applies there too.
func TestInsertExperiment_MetricsGatedByRecorder(t *testing.T) {
	method := "InsertExperiment"
	newExperiment := func() *_pubsub.Experiment {
		return &_pubsub.Experiment{
			Id:        1,
			ProjectId: 1,
			Status:    _pubsub.Experiment_Active,
			StartTime: timestamppb.New(time.Now()),
			EndTime:   timestamppb.New(time.Now().Add(time.Hour)),
		}
	}

	t.Run("no recorder by default", func(t *testing.T) {
		s := &LocalStorage{Experiments: map[ProjectId][]*ExperimentIndex{}}
		s.InsertExperiment(newExperiment())
	})

	t.Run("recorder set", func(t *testing.T) {
		recorder := newFakeMetricsRecorder()
		s := &LocalStorage{Experiments: map[ProjectId][]*ExperimentIndex{}}
		s.SetMetricsRecorder(recorder)

		s.InsertExperiment(newExperiment())

		require.Equal(t, 1, recorder.callCounts[method])
		require.Equal(t, 1, recorder.durationCounts[method])
	})
}
