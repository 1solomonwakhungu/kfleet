package hubcontact

import (
	"sync"
	"testing"
	"time"
)

func TestTrackerReadyLifecycle(t *testing.T) {
	tracker := NewTracker(45 * time.Second)
	clock := time.Now()
	tracker.SetClockForTesting(func() time.Time { return clock })

	if tracker.Ready() {
		t.Fatal("Ready() = true before any successful registration, want false")
	}

	tracker.MarkRegistered()
	if !tracker.Ready() {
		t.Fatal("Ready() = false right after a successful registration, want true")
	}

	// Inside the staleness window (3x the heartbeat interval) the agent stays ready.
	clock = clock.Add(44 * time.Second)
	if !tracker.Ready() {
		t.Fatal("Ready() = false 44s after contact with a 45s window, want true")
	}

	// Past the window without new contact the agent is no longer ready.
	clock = clock.Add(2 * time.Second)
	if tracker.Ready() {
		t.Fatal("Ready() = true 46s after contact with a 45s window, want false")
	}

	// A heartbeat or status report refreshes the contact timestamp.
	tracker.MarkContact()
	if !tracker.Ready() {
		t.Fatal("Ready() = false right after a successful heartbeat, want true")
	}
}

func TestTrackerNilIsNeverReady(t *testing.T) {
	var tracker *Tracker
	if tracker.Ready() {
		t.Fatal("Ready() on a nil tracker = true, want false")
	}
	tracker.MarkRegistered()
	tracker.MarkContact()
	tracker.SetClockForTesting(time.Now)
	if tracker.Ready() {
		t.Fatal("Ready() on a nil tracker = true after marking contact, want false")
	}
}

func TestTrackerConcurrentUse(_ *testing.T) {
	tracker := NewTracker(time.Second)
	tracker.MarkRegistered()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				tracker.MarkContact()
				_ = tracker.Ready()
			}
		}()
	}
	wg.Wait()
}
