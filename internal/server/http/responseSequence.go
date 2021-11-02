package http

import (
	"sync"
	"time"

	"github.com/cheddartv/mockarena/internal/server/http/config"
)

type responseSequence struct {
	responses []*config.Response

	responseCounts       map[*config.Response]uint
	currentResponseStart *time.Time

	sync.RWMutex
}

func newResponseSequence(responses []*config.Response) *responseSequence {
	return &responseSequence{
		responses: responses,
	}
}

func (rs *responseSequence) next() *config.Response {
	rs.Lock()
	defer rs.Unlock()

	return rs._next()
}

func (rs *responseSequence) _next() *config.Response {
	if len(rs.responses) == 0 {
		return nil
	}

	var (
		response = rs.responses[0]
		repeat   = response.Repeat
	)

	switch {
	case 0 < repeat.Count:
		if rs.responseCounts == nil {
			rs.responseCounts = make(map[*config.Response]uint)
		}

		if repeat.Count <= rs.responseCounts[response] {
			// response count limit exceeded, discard and move on
			rs.responses = rs.responses[1:]
			return rs._next()
		}

		rs.responseCounts[response]++

		return response
	case !repeat.Until.IsZero():
		var cutOff = repeat.Until

		if !time.Now().Before(cutOff) {
			// time limit exceeded, discard and move on
			rs.responses = rs.responses[1:]
			return rs._next()
		}

		return response
	case 0 < repeat.For:
		var now = time.Now()

		if rs.currentResponseStart == nil {
			rs.currentResponseStart = &now
		}

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
