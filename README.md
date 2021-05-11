Clock
=====

[![Go](https://github.com/kraney/clock/actions/workflows/go.yml/badge.svg)](https://github.com/kraney/clock/actions/workflows/go.yml)

Clock is a small library for mocking time in Go. It provides an interface
around the standard library's [`time`][time] package so that the application
can use the realtime clock while tests can use the mock clock.

This package is derived from github.com/benbjohnson/clock, but has been modified
to support better synchronization with separate threads that start timers

[time]: http://golang.org/pkg/time/


## Usage

### Realtime Clock

Your application can maintain a `Clock` variable that will allow realtime and
mock clocks to be interchangeable. For example, if you had an `Application` type:

```go
import "github.com/kraney/clock"

type Application struct {
	Clock clock.Clock
}
```

You could initialize it to use the realtime clock like this:

```go
var app Application
app.Clock = clock.New()
...
```

Then all timers and time-related functionality should be performed from the
`Clock` variable.

Alternately, you can use the functions in the clock package that mimic Go's time package.


### Mocking time

In your tests, you will want to use a `Mock` clock:

```go
import (
	"testing"

	"github.com/kraney/clock"
)

func TestApplication_DoSomething(t *testing.T) {
	mock := clock.NewMock(t, 0)
	app := Application{Clock: mock}
	...
}
```

Or, if you're using the functions in the clock package, you must override
the system clock

```go
func TestApplication_DoSomething(t *testing.T) {
  clock.SetSystemClock(clock.NewMock(t, 0))
  ...
}
```

Now that you've initialized your application to use the mock clock, you can
adjust the time programmatically. The mock clock always starts from the Unix
epoch (midnight UTC on Jan 1, 1970).

### Synchronization

One tricky part of mocking time is that generally, the events you are trying to control are
in a separate gothread. This creates a couple of common race conditions:
 * You need to be confident another thread has called NewTicker or NewTimer or
   Sleep _before_ you advance the clock. Otherwise their wait time will start
   _after_ you're done moving the clock, so it'll never be satisfied
 * You need to be confident the thread that handles the timer is _done_ doing its thing
   so you can then do asserts about what the result was.

This library provides some tools to make this easier to deal with.

#### Expect

The mock provides the ability to "expect" a specific number of timer starts or a specific number 
of processed time events. You prepare this by calling Expect* or setting an expect option _before_ 
the other threads will be doing this. Then, when you need to be sure those threads are done, you can
call Wait* which will block until the expected count is reached.

Optionally, you can toggle the mock to fail a test if an unexpected event
happens using the FailOnUnexpectedUpcomingEvent option. Once this is set, any
new timers that aren't accounted for by a call to Expect will fail a test. This
behavior continues on all subsequent calls unless you expressly turn it back
off using IgnoreUnexpectedUpcomingEvent.

The helpers used for synchronization are also exposed for use, such as to wait for a thread
to handle a timer during a test. The are similar to sync.WaitGroup, but
 * OptionalCheckpoint does not panic on Done() for unexpected calls
 * FailOnUnexpectedCheckpoint will fail a test (rather than panic) on unexpected calls to Done()

### Defaults

The mock returned by `NewMock` assumes / enforces
 * that tests should fail when unexpected timer events happen (if testing.T is not nil)
 * that clock should block until all expected timers are started before advancing the clock

It's expected that this is usually (always?) the right approach during testing. If there is
a use case where it's not, then `NewUnconfirmedMock()` can be used instead. It supports all
of the same synchronization features but does not enforce them by default, leaving it to the
user to choose when to specify a Wait or to turn on FailOnUnexpectedEvent.

### Controlling time

The mock clock provides the same functions that the standard library's `time`
package provides. For example, to find the current time, you use the `Now()`
function:

```go
mock := clock.NewMock()

// Find the current time.
mock.Now().UTC() // 1970-01-01 00:00:00 +0000 UTC

// Move the clock forward.
mock.Add(2 * time.Hour)

// Check the time again. It's 2 hours later!
mock.Now().UTC() // 1970-01-01 02:00:00 +0000 UTC
```

Timers and Tickers are also controlled by this same mock clock. They will only
execute when the clock is moved forward:

```go
mock := clock.NewMock(clock.ExpectUpcomingStarts(1), clock.FailOnUnexpectedUpcomingEvent(t))
count := 0
confirm := clock.NewOptionalCheckPoint(CheckpointName("incremented"))

// Kick off a timer to increment every 1 mock second.
go func() {
    ticker := mock.Ticker(1 * time.Second)
    for {
        <-ticker.C
        count++
        // this optional call provides synchronization that the timer event has been handled
        confirm.Done()
    }
}()

// Wait for all expected starts, then move the clock forward 10 seconds.
// Expect a confirm. After advancing the clock, wait until the confirm has been seen
confirm.Add(2)
mock.Add(10 * time.Second, clock.WaitBefore)

// this will ensure this thread waits until the timer thread has defintely run and handled the timer event
confirm.Wait()

// This prints 10.
fmt.Println(count)

// for convenience and readability, you can pass options to make waits happen
confirm.Add(2)
mock.Add(20 * time.Second)
confirm.Wait()

// This prints 30.
fmt.Println(count)
```
