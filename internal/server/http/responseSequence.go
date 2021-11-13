package http

import (
	"sync"
	"time"
)

type responseSequence struct {
	responses []*Response

	responseCounts       map[*Response]uint
	currentResponseStart *time.Time

	doneChan chan struct{}
	done     bool

	runningUntilTimer   bool
	runningUntilTimerMu sync.Mutex
	runningForTimer     bool
	runningForTimerMu   sync.Mutex

	sync.RWMutex
}

func newResponseSequence(responses []*Response) *responseSequence {
	var rs = responseSequence{
		responses:      responses,
		responseCounts: make(map[*Response]uint),
		doneChan:       make(chan struct{}),
	}

	if len(rs.responses) == 1 {
		var r = rs.responses[0]
		if r.Repeat.Forever == "nonblocking" {
			go func() {
				rs.done = true
				rs.doneChan <- struct{}{}
			}()
		}
	}

	return &rs
}

func (rs *responseSequence) next() *Response {
	rs.Lock()
	defer rs.Unlock()

	return rs._next()
}

func (rs *responseSequence) _next() *Response {
	if len(rs.responses) == 0 {
		return nil
	}

	var (
		response = rs.responses[0]
		repeat   = response.Repeat

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

		rs.responseCounts[response]++

		return response
	case repeat.Forever == "blocking":
		rs.responseCounts[response]++

		return response
	case 0 < repeat.Count:
		if rs.responseCounts == nil {
			rs.responseCounts = make(map[*Response]uint)
		}

		if repeat.Count <= rs.responseCounts[response] {
			// response count limit exceeded, discard and move on
			rs.responses = rs.responses[1:]
			return rs._next()
		}

		if repeat.Count-1 <= rs.responseCounts[response] {
			go done()
		}

		rs.responseCounts[response]++

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
