package bit

import (
	"fmt"
	"math"
	"sync/atomic"
	"time"
)

type DurationMetric struct {
	Count         uint64
	TotalDuration time.Duration
}

func WithDurationMetric(m *DurationMetric, f func()) {
	start := time.Now()
	f()
	m.PutEvent(time.Since(start))
}

func WithDurationMetricResult[T any](m *DurationMetric, f func() T) T {
	var result T
	WithDurationMetric(m, func() {
		result = f()
	})
	return result
}

func (m *DurationMetric) PutEvent(d time.Duration) {
	atomic.AddUint64(&m.Count, 1)
	atomic.AddInt64((*int64)(&m.TotalDuration), int64(d))
}

func (m *DurationMetric) Load() DurationMetric {
	// Atomically load metrics field by field.
	return DurationMetric{
		Count:         atomic.LoadUint64(&m.Count),
		TotalDuration: time.Duration(atomic.LoadInt64((*int64)(&m.TotalDuration))),
	}
}

func (m DurationMetric) AverageDuration() time.Duration {
	if m.Count == 0 {
		return 0
	}
	return time.Duration(float64(m.TotalDuration) / float64(m.Count))
}

func (m DurationMetric) AverageFPS() float64 {
	if m.TotalDuration == 0 {
		return math.Inf(1)
	}
	return float64(m.Count) / m.TotalDuration.Seconds()
}

func (m DurationMetric) String() string {
	var totalDurationStr string
	if m.TotalDuration < time.Second {
		totalDurationStr = "<1s"
	} else {
		totalDurationStr = m.TotalDuration.Round(time.Second).String()
	}
	return fmt.Sprintf(
		"[%d ops, %v total time, %v avg time]",
		m.Count,
		totalDurationStr,
		m.AverageDuration().Round(time.Nanosecond),
	)
}
