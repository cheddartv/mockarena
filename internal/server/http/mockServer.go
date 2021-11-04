package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cheddartv/mockarena/internal/server/http/config"
)

type MockServer struct {
	stats    httpServerStats
	sequence *responseSequence
	conf     config.ServerConfig

	sync.Mutex
}

func NewMockServer(c config.ServerConfig) *MockServer {
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
		defer func() {
			io.Copy(w, &pw)
		}()
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

type ProxyResponseWriter struct {
	header http.Header
	status int
	closed bool
	bytes.Buffer
}

func (pw *ProxyResponseWriter) Header() http.Header {
	return pw.header
}

func (pw *ProxyResponseWriter) WriteHeader(statusCode int) {
	pw.status = statusCode
}

func (pw *ProxyResponseWriter) Write(p []byte) (int, error) {
	if pw.closed {
		return len(p), nil
	}

	return pw.Buffer.Write(p)
}
