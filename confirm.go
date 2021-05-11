package clock

var (
	WaitForConfirmsBefore = &WaitForConfirmsBeforeOption{}
	WaitForConfirmsAfter  = &WaitForConfirmsAfterOption{}
)

// Confirmable is an interface that provides extra functionality in support
// of synchronization for tests. If your timer handler does not have a natural message
// or channel to wait on, the Confirm method provides a way to do so in the test that
// is a no-op in production use.
type Confirmable interface {
	Confirm()
}

type ExpectUpcomingConfirmsOption struct {
	confirms int
}

func ExpectUpcomingConfirms(confirms int) *ExpectUpcomingConfirmsOption {
	return &ExpectUpcomingConfirmsOption{confirms}
}

func (o *ExpectUpcomingConfirmsOption) PriorEventsOption(mock *UnsynchronizedMock) {}

func (o *ExpectUpcomingConfirmsOption) UpcomingEventsOption(mock *UnsynchronizedMock) {
	mock.ExpectConfirms(int(o.confirms))
}

func (o *ExpectUpcomingConfirmsOption) AfterClockAdvanceOption(mock *UnsynchronizedMock) {}

type WaitForConfirmsBeforeOption struct{}

func (o *WaitForConfirmsBeforeOption) PriorEventsOption(mock *UnsynchronizedMock) {
	mock.WaitForConfirm()
}

func (o *WaitForConfirmsBeforeOption) UpcomingEventsOption(mock *UnsynchronizedMock) {}

func (o *WaitForConfirmsBeforeOption) AfterClockAdvanceOption(mock *UnsynchronizedMock) {}

type WaitForConfirmsAfterOption struct{}

func (o *WaitForConfirmsAfterOption) PriorEventsOption(mock *UnsynchronizedMock) {}

func (o *WaitForConfirmsAfterOption) UpcomingEventsOption(mock *UnsynchronizedMock) {}

func (o *WaitForConfirmsAfterOption) AfterClockAdvanceOption(mock *UnsynchronizedMock) {
	mock.WaitForConfirm()
}

func Confirm() {
	if cfrm, ok := systemClock.(Confirmable); ok {
		cfrm.Confirm()
	}
}

// Confirm for a system clock is a no-op
func (c *clock) Confirm() {}

// Confirm confirms that a timer event has been processed - no op for system clock, but allows synchronization of the mock
func (t *Timer) Confirm() {
	if t.timer != nil {
		return
	}

	if cfrm, ok := MockableClock(t.mock).(Confirmable); ok {
		cfrm.Confirm()
	}
}

// Confirm confirms that a ticker event has been processed - no op for system clock, but allows synchronization of the mock
func (t *Ticker) Confirm() {
	if t.ticker != nil {
		return
	}

	if cfrm, ok := MockableClock(t.mock).(Confirmable); ok {
		cfrm.Confirm()
	}
}

// ExpectConfirms informs the mock how many timers should have been confirmed before we advance the clock
func (m *UnsynchronizedMock) ExpectConfirms(confirmCount int) {
	m.mu.Lock()
	sp := m.syncPoints[OnConfirm]
	m.mu.Unlock()
	sp.Add(confirmCount)
}

// WaitForConfirm will block until all expected timers have been confirmed
func (m *UnsynchronizedMock) WaitForConfirm() {
	m.mu.Lock()
	sp := m.syncPoints[OnConfirm]
	m.mu.Unlock()
	sp.Wait()
}

func (m *UnsynchronizedMock) Confirm() {
	m.mu.Lock()
	sp := m.syncPoints[OnConfirm]
	m.mu.Unlock()
	sp.Done()
}
