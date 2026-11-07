package agent

import "testing"

// ING-F2: the agent's 15s replay ticker resends anything still marked
// unprocessed in the DB, which includes a log sent moments ago that is
// simply waiting on a slightly slow ack — not lost. inFlight is what lets
// replayUnprocessed tell the two apart.
func TestInFlightTracksSentButUnacked(t *testing.T) {
	l := &LogProcessor{}

	if l.isInFlight("a") {
		t.Fatal("a log that was never sent must not be in flight")
	}

	l.markInFlight("a")
	if !l.isInFlight("a") {
		t.Fatal("a log just sent must be in flight until it's acked")
	}

	l.clearInFlight("a")
	if l.isInFlight("a") {
		t.Fatal("a log must stop being in flight once its ack arrives")
	}
}

// On reconnect, whatever was in flight on the dead connection will never get
// an ack on it — resetInFlight must make those eligible for replay again.
func TestResetInFlightClearsEverythingOnReconnect(t *testing.T) {
	l := &LogProcessor{}
	l.markInFlight("a")
	l.markInFlight("b")

	l.resetInFlight()

	if l.isInFlight("a") || l.isInFlight("b") {
		t.Fatal("resetInFlight must clear every tracked id for the new connection")
	}
}
