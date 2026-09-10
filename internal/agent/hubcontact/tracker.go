// Package hubcontact tracks the agent's most recent successful contact with
// the hub so the readiness probe can report honest reachability.
package hubcontact

import (
	"sync"
	"time"
)

// Tracker records when the agent last reached the hub successfully. The
// registrar marks registrations and heartbeats, the reporter marks status
// reports. Ready reports whether the agent has ever registered successfully
// and the most recent success is within the staleness window.
//
// A nil Tracker is usable: Ready always reports false, and the Mark methods
// are no-ops, so components may hold a nil tracker when tracking is disabled.
type Tracker struct {
	mu          sync.Mutex
	registered  bool
	lastContact time.Time
	staleness   time.Duration
	now         func() time.Time
}

// NewTracker returns a tracker that considers contact stale after staleness.
func NewTracker(staleness time.Duration) *Tracker {
	return &Tracker{
		staleness: staleness,
		now:       time.Now,
	}
}

// MarkRegistered records a successful registration with the hub. The
// registration itself counts as successful contact.
func (t *Tracker) MarkRegistered() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.registered = true
	t.lastContact = t.now()
}

// MarkContact records a successful non-registration hub call, such as a
// heartbeat or a status report.
func (t *Tracker) MarkContact() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastContact = t.now()
}

// Ready reports whether the agent has ever registered successfully and the
// last successful hub contact is within the staleness window.
func (t *Tracker) Ready() bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.registered {
		return false
	}
	return t.now().Sub(t.lastContact) <= t.staleness
}

// SetClockForTesting replaces the tracker's clock. It is meant for tests only.
func (t *Tracker) SetClockForTesting(now func() time.Time) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.now = now
}
