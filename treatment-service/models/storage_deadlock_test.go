package models

import (
	"sync"
	"testing"
	"time"

	_pubsub "github.com/caraml-dev/xp/common/pubsub"
)

// TestFindSubscribedProjectSettingsById_NestedRLockDeadlocksWithPendingWriter reproduces the
// root cause suspected for the reported GetTreatmentForRequest deadlock: findSubscribedProjectSettingsById
// takes s.RLock() and then, while still holding it, calls findProjectSettingsById which takes
// s.RLock() again on the same goroutine.
//
// A recursive RLock is fine in isolation, but Go's sync.RWMutex blocks *new* readers once a
// writer is queued (to prevent writer starvation, see sync.RWMutex docs). So if a writer
// (e.g. PollerService.Refresh -> Init(), or a pubsub InsertExperiment/UpdateProjectSettings
// handler calling s.Lock()) queues up between the outer and the nested RLock, the nested RLock
// blocks forever waiting for the writer, and the writer blocks forever waiting for the outer
// RLock to be released (which never happens, since the function holding it is stuck in the
// nested call). That's a genuine circular-wait deadlock, not just contention.
//
// This test reproduces that interleaving deterministically: it manually holds the outer RLock
// (standing in for findSubscribedProjectSettingsById's lock), starts a writer that queues on
// s.Lock(), and then invokes the real findProjectSettingsById -- the exact call that is nested
// inside findSubscribedProjectSettingsById in production code.
func TestFindSubscribedProjectSettingsById_NestedRLockDeadlocksWithPendingWriter(t *testing.T) {
	s := &LocalStorage{
		ProjectSettings: []*_pubsub.ProjectSettings{
			{ProjectId: 1},
		},
	}

	// Simulate findSubscribedProjectSettingsById's outer RLock, already held on this goroutine.
	s.RLock()

	// Start a writer that mirrors PollerService.Refresh -> Init(), or a pubsub update handler.
	// It queues on Lock() while the outer RLock above is held.
	writerDone := make(chan struct{})
	go func() {
		s.Lock()
		s.Unlock()
		close(writerDone)
	}()

	// Give the writer time to actually queue behind the held RLock. This is inherently timing
	// based, but the window only needs to be wide enough for the writer's Lock() call to enter
	// its wait queue -- not for anything to complete -- so it is not flaky in practice.
	time.Sleep(200 * time.Millisecond)

	// This is the exact nested call findSubscribedProjectSettingsById makes while still holding
	// the outer RLock. With the writer now queued, this must block forever.
	nestedRLockDone := make(chan *_pubsub.ProjectSettings, 1)
	go func() {
		nestedRLockDone <- s.findProjectSettingsById(1)
	}()

	select {
	case <-nestedRLockDone:
		s.RUnlock()
		t.Fatal("expected the nested RLock to deadlock while a writer was queued, but it completed instead; " +
			"the suspected root cause did not reproduce")
	case <-writerDone:
		s.RUnlock()
		t.Fatal("expected the writer to be blocked behind the still-held outer RLock, but it completed instead; " +
			"the suspected root cause did not reproduce")
	case <-time.After(2 * time.Second):
		// Deadlock reproduced: neither the nested reader nor the queued writer could make
		// progress. Root cause confirmed.
	}

	// Release the outer RLock so the goroutines above (which are genuinely deadlocked in
	// production) can unwind and this test doesn't leak them past its own completion.
	s.RUnlock()
	<-writerDone
	<-nestedRLockDone
}

// TestFindProjectSettingsWithId_ConcurrentWithWriter_NoDeadlock exercises the real production
// call graph (FindProjectSettingsWithId -> findSubscribedProjectSettingsById -> findProjectSettingsById)
// under concurrent read/write load, to prove the nested-RLock hazard is actually gone from that
// path (not just demonstrated in isolation). Many reader goroutines keep the nested-call path hot
// while a writer goroutine repeatedly takes s.Lock(), so if the nested RLock ever reappears, a
// writer will eventually queue between the outer and inner RLock calls on some reader goroutine
// and the whole test hangs past the timeout.
func TestFindProjectSettingsWithId_ConcurrentWithWriter_NoDeadlock(t *testing.T) {
	projectId := ProjectId(1)
	s := &LocalStorage{
		subscribedProjectIds: []ProjectId{projectId},
		ProjectSettings:      []*_pubsub.ProjectSettings{{ProjectId: int64(projectId)}},
	}

	stop := make(chan struct{})
	var readers sync.WaitGroup
	for i := 0; i < 50; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					s.FindProjectSettingsWithId(projectId)
				}
			}
		}()
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for i := 0; i < 2000; i++ {
			s.UpdateProjectSettings(&_pubsub.ProjectSettings{ProjectId: int64(projectId)})
		}
	}()

	select {
	case <-writerDone:
		// Writer completed all its Lock/Unlock cycles without ever being stuck behind a
		// reader that recursively RLocks -- no deadlock.
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: writer never completed while concurrent readers were calling " +
			"FindProjectSettingsWithId, indicating a nested RLock on the read path")
	}

	close(stop)
	readers.Wait()
}
