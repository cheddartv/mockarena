package mysql

import (
	"sync"
	"time"
)

type responseSequence struct {
	responses []*RepeatableResponse

	responseCounts       map[*RepeatableResponse]uint
	currentResponseStart *time.Time

	doneChan chan struct{}
	done     bool

	runningUntilTimer   bool
	runningUntilTimerMu sync.Mutex
	runningForTimer     bool
	runningForTimerMu   sync.Mutex

	sync.RWMutex
}

func newResponseSequence(responses []*RepeatableResponse) *responseSequence {
	var rs = responseSequence{
		responses: responses,
		doneChan:  make(chan struct{}),
	}

	return &rs
}

func (rs *responseSequence) next() Response {
	rs.Lock()
	defer rs.Unlock()

	return rs._next()
}

func (rs *responseSequence) _next() Response {
	if len(rs.responses) == 0 {
		return nil
	}

	var (
		rr       = rs.responses[0]
		response = rr.Response
		repeat   = rr.Repeat

		done = func() {
			if len(rs.responses) <= 1 {
				rs.done = true
				rs.doneChan <- struct{}{}
			}
		}
	)

	switch {
	case repeat.Forever == "nonblocking":
		if !rs.done {
			go done()
		}

		rs.responseCounts[rr]++

		return response
	case repeat.Forever == "blocking":
		rs.responseCounts[rr]++

		return response
	case 0 < repeat.Count:
		if rs.responseCounts == nil {
			rs.responseCounts = make(map[*RepeatableResponse]uint)
		}

		if repeat.Count <= rs.responseCounts[rr] {
			// response count limit exceeded, discard and move on
			rs.responses = rs.responses[1:]
			return rs._next()
		}

		if repeat.Count-1 <= rs.responseCounts[rr] {
			go done()
		}

		rs.responseCounts[rr]++

		return response
	case !repeat.Until.IsZero():
		var (
			cutOff = repeat.Until
			now    = time.Now()
		)

		if !now.Before(cutOff) {
			// time limit exceeded, discard and move on
			rs.responses = rs.responses[1:]
			return rs._next()
		}

		rs.runningUntilTimerMu.Lock()
		if !rs.runningUntilTimer {
			rs.runningUntilTimer = true

			time.AfterFunc(cutOff.Sub(now), func() {
				rs.runningUntilTimerMu.Lock()
				rs.runningUntilTimer = false
				rs.runningUntilTimerMu.Unlock()
				done()
			})
		}
		rs.runningUntilTimerMu.Unlock()

		return response
	case 0 < repeat.For:
		var now = time.Now()

		if rs.currentResponseStart == nil {
			rs.currentResponseStart = &now
		}

		rs.runningForTimerMu.Lock()
		if !rs.runningForTimer {
			rs.runningForTimer = true

			time.AfterFunc(repeat.For, func() {
				rs.runningForTimerMu.Lock()
				rs.runningForTimer = false
				rs.runningForTimerMu.Unlock()
				done()
			})
		}
		rs.runningForTimerMu.Unlock()

		if repeat.For < time.Since(*rs.currentResponseStart) {
			// time limit exceeded, discard and move on
			rs.currentResponseStart = nil

			rs.responses = rs.responses[1:]
			return rs._next()
		}

		return response
	}

	return nil
}
