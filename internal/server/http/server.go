package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cheddartv/mockarena/internal/server/http/config"
)

type HTTPServer struct {
	stats    httpServerStats
	sequence *responseSequence
	conf     config.ServerConfig

	sync.Mutex
}

func NewHTTPServer(c config.ServerConfig) *HTTPServer {
	var s HTTPServer

	s.conf = c
	s.sequence = newResponseSequence(c.ReturnSequence)

	return &s
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
