package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type HTTPServerStats struct {
	Count      uint
	PathCounts map[string]uint

	sync.Mutex
}

func (ss *HTTPServerStats) Inc(path string) {
	ss.Lock()
	ss.Count++
	if ss.PathCounts == nil {
		ss.PathCounts = make(map[string]uint)
	}
	ss.PathCounts[path]++
	ss.Unlock()
}

type HTTPServer struct {
	stats HTTPServerStats

	Sequential bool
	Sequence   ResponseSequence

	sync.Mutex
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Sequential {
		s.Lock()
		defer s.Unlock()
	}

	s.stats.Inc(r.URL.Path)

	{
		var data, err = json.MarshalIndent(s.stats, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(data))
	}

	var response = s.nextResponse()

	if d := response.Delay; 0 < d {
		time.Sleep(d)
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

func (s *HTTPServer) nextResponse() *Response {
	return s.Sequence.next()
}
