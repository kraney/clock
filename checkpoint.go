package clock

import (
	"fmt"
	"sync"
	"testing"
)

type CheckpointName string

type Checkpoint interface {
	Add(delta int)
	Done()
	Wait()
}

var (
	checkpoints map[CheckpointName]Checkpoint
)

func init() {
	checkpoints = make(map[CheckpointName]Checkpoint)
}

func Done(name CheckpointName) {
	cp, ok := checkpoints[name]
	if ok {
		cp.Done()
	}
}

func EnableCheckpoint(name CheckpointName, cp Checkpoint) error {
	_, ok := checkpoints[name]
	if ok {
		return fmt.Errorf("Checkpoint %v already exists", name)
	}
	checkpoints[name] = cp
	return nil
}

func DisableCheckpoint(name CheckpointName) {
	delete(checkpoints, name)
}

// OptionalCheckpoint provides waitgroup-like functionality with assistance
// to avoid races when setting up new waits for already-running threads.
type OptionalCheckpoint struct {
	name        CheckpointName
	outstanding chan int
}

func NewOptionalCheckPoint(name CheckpointName) *OptionalCheckpoint {
	ret := &OptionalCheckpoint{
		name:        name,
		outstanding: make(chan int, 1),
	}
	ret.outstanding <- 0
	return ret
}

func (s *OptionalCheckpoint) Add(delta int) {
	s.updateOutstanding(delta)
}

func (s *OptionalCheckpoint) Done() {
	s.updateOutstanding(-1)
}

func (s *OptionalCheckpoint) Wait() {
	os := <-s.outstanding
	for os > 0 {
		update := <-s.outstanding
		os += update
	}
	s.outstanding <- 0
}

func (s *OptionalCheckpoint) String() string {
	return string(s.name)
}

func (s *OptionalCheckpoint) updateOutstanding(delta int) {
	for {
		select {
		case s.outstanding <- delta:
			// we were able to write to the channel, so we're done
			return
		default:
			// there's an outstanding value in the channel, update from it and retry
			os := <-s.outstanding
			delta += os
		}
	}
}

// FailOnUnexpectedCheckpoint extends SimpleSyncPoint so that excess calls do Done fail a
// test.
type FailOnUnexpectedCheckpoint struct {
	name     CheckpointName
	mu       sync.Mutex
	wg       sync.WaitGroup
	expected int
	t        *testing.T
}

func NewFailOnUnexpectedCheckpoint(name CheckpointName, t *testing.T) *FailOnUnexpectedCheckpoint {
	return &FailOnUnexpectedCheckpoint{
		name: name,
		mu:   sync.Mutex{},
		wg:   sync.WaitGroup{},
		t:    t,
	}
}

func (t *FailOnUnexpectedCheckpoint) Add(delta int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.expected = t.expected + delta
	t.wg.Add(delta)
}

func (t *FailOnUnexpectedCheckpoint) Done() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.expected <= 0 {
		t.t.Helper()
		t.t.Errorf("unexpected %v event", t.name)
		return
	}
	t.expected--
	t.wg.Done()
}

func (t *FailOnUnexpectedCheckpoint) Wait() {
	t.wg.Wait()
	t.expected = 0
}

func (t *FailOnUnexpectedCheckpoint) String() string {
	return string(t.name)
}
