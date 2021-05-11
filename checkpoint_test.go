package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testCheckpoint CheckpointName = "TestCheckpoint"

func TestOptionalCheckPoint(t *testing.T) {
	cp := NewOptionalCheckPoint(testCheckpoint)

	// Done with no adds should not panic
	cp.Done()

	// cp.Wait with no adds should return immediately
	cp.Wait()

	// Add then done should return immediately
	cp.Add(1)
	cp.Done()
	cp.Wait()

	// Too many Done should not panic
	cp.Add(1)
	cp.Done()
	cp.Done()
	cp.Wait()

	// Done then add should also return immediately
	cp.Done()
	cp.Add(1)
	cp.Wait()

	// Done then add then done should also return immediately
	cp.Done()
	cp.Add(2)
	cp.Done()
	cp.Wait()

	cp.Add(1)
	var called bool
	go func() {
		time.Sleep(50 * time.Millisecond)
		called = true
		cp.Done()
	}()
	cp.Wait()
	assert.True(t, called, "wait did not block")
}

func TestFailOnUnexpectedCheckpoint(t *testing.T) {
	experiment := &testing.T{}
	cp := NewFailOnUnexpectedCheckpoint(testCheckpoint, experiment)

	// Done with no adds should cause a failure
	cp.Done()
	assert.True(t, experiment.Failed(), "lack of failure on unexpected")

	// cp.Wait with no adds should return immediately
	cp.Wait()

	experiment = &testing.T{}
	cp = NewFailOnUnexpectedCheckpoint(testCheckpoint, experiment)
	// Too many Done should fail
	cp.Add(1)
	cp.Done()
	cp.Done()
	assert.True(t, experiment.Failed(), "lack of failure on unexpected")
	cp.Wait()

	experiment = &testing.T{}
	cp = NewFailOnUnexpectedCheckpoint(testCheckpoint, experiment)
	// Add then done should return immediately
	cp.Add(1)
	cp.Done()
	cp.Wait()
	assert.False(t, experiment.Failed(), "failure without unexpected")

	// Multiple done should return immediately
	cp.Add(2)
	cp.Done()
	cp.Done()
	cp.Wait()
	assert.False(t, experiment.Failed(), "failure without unexpected")

	cp.Add(1)
	var called bool
	go func() {
		time.Sleep(50 * time.Millisecond)
		called = true
		cp.Done()
	}()
	cp.Wait()
	assert.Equal(t, true, called, "wait did not block")
	assert.False(t, experiment.Failed(), "failure without unexpected")
}
