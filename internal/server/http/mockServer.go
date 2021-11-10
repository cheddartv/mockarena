package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type MockServer struct {
	stats    httpServerStats
	sequence *responseSequence
	conf     ServerConfig

	sync.Mutex
}

func NewMockServer(c ServerConfig) *MockServer {
	var s MockServer

	s.conf = c
	s.sequence = newResponseSequence(c.ReturnSequence)

	return &s
}

func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.conf.Serial {
		s.Lock()
		defer s.Unlock()
	}

	s.stats.inc(r.URL.Path)

	{
		var data, err = json.MarshalIndent(s.stats, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(data))
	}

	var response = s.sequence.next()

	if d := response.Delay; 0 < d {
		time.Sleep(d)
	}

	if response.Proxy != nil {
		var pw ProxyResponseWriter
		pw.header = make(http.Header)
		response.Proxy.ServeHTTP(&pw, r)
		pw.closed = true
		defer func(w http.ResponseWriter) {
			var header = w.Header()
			for k, v := range pw.header {
				for _, h := range v {
					header.Add(k, h)
				}
			}
			w.WriteHeader(pw.status)
			io.Copy(w, &pw)
		}(w)
		w = &pw
	}

	if response.Header != nil {
		var header = w.Header()
		for _, h := range response.Header {
			header.Add(h.Key, h.Value)
		}
	}

	if c := response.Status; 0 < c {
		w.WriteHeader(c)
	}

	if b := response.Body; b != nil {
		w.Write(b)
	}
}

type responseSequence struct {
	responses []*Response

	responseCounts       map[*Response]uint
	currentResponseStart *time.Time

	sync.RWMutex
}

func newResponseSequence(responses []*Response) *responseSequence {
	return &responseSequence{
		responses: responses,
	}
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
	)

	switch {
	case 0 < repeat.Count:
		if rs.responseCounts == nil {
			rs.responseCounts = make(map[*Response]uint)
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
