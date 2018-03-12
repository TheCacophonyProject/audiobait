// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package window

import (
	"time"
)

// New creates a Window instance which represents a recurring window
// between two times of day. If `start` is after `end` then the time
// window is assumed to cross over midnight. If `start` and `end` are
// the same then the window is always active.
func New(start, end time.Time) *Window {
	start = normaliseTime(start)
	end = normaliseTime(end)

	xMidnight := false
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
		xMidnight = true
	}

	return &Window{
		Start:     start,
		End:       end,
		Now:       time.Now,
		xMidnight: xMidnight,
	}
}

// Window represents a recurring window between two times of day.
// The Now field can be use to override the time source (for testing).
type Window struct {
	Start     time.Time
	End       time.Time
	Now       func() time.Time
	xMidnight bool
}

// Active returns true if the time window is currently active.
func (w *Window) Active() bool {
	return w.Until() == time.Duration(0)
}

// Until returns the duration until the next time window starts.
func (w *Window) Until() time.Duration {
	if w.Start == w.End {
		return time.Duration(0)
	}

	now := normaliseTime(w.Now())
	if w.xMidnight && now.Before(w.Start) {
		now = now.Add(24 * time.Hour)
	}

	untilStart := w.Start.Sub(now)
	if untilStart > 0 {
		// Before window start.
		return untilStart
	}
	if w.End.Sub(now) >= 0 {
		// During window.
		return time.Duration(0)
	}
	// After window.
	return w.Start.Add(24 * time.Hour).Sub(now)
}

func normaliseTime(t time.Time) time.Time {
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}
