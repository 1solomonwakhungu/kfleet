package registrar

import (
	"testing"
	"time"
)

func TestBackoffNextGrowthAndCap(t *testing.T) {
	tests := []struct {
		name string
		max  time.Duration
		want []time.Duration
	}{
		{
			name: "exponential growth",
			max:  time.Minute,
			want: []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second},
		},
		{
			name: "maximum cap",
			max:  4 * time.Second,
			want: []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 4 * time.Second},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backoff := &Backoff{
				Base:        time.Second,
				Max:         test.max,
				factor:      2,
				randFloat64: func() float64 { return 0.5 },
			}
			for index, want := range test.want {
				if got := backoff.Next(); got != want {
					t.Errorf("Next() call %d = %v, want %v", index+1, got, want)
				}
			}
		})
	}
}

func TestBackoffNextJitterBounds(t *testing.T) {
	tests := []struct {
		name      string
		randValue float64
		want      time.Duration
	}{
		{name: "lower bound", randValue: 0, want: 800 * time.Millisecond},
		{name: "no jitter", randValue: 0.5, want: time.Second},
		{name: "upper bound", randValue: 1, want: 1200 * time.Millisecond},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backoff := NewBackoff()
			backoff.randFloat64 = func() float64 { return test.randValue }
			if got := backoff.Next(); got != test.want {
				t.Fatalf("Next() = %v, want %v", got, test.want)
			}
		})
	}
}
