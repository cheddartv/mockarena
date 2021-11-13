package http

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

var allMethods = []string{
	"GET",
	"HEAD",
	"POST",
	"PUT",
	"DELETE",
	"CONNECT",
	"OPTIONS",
	"TRACE",
}

type MockServer struct {
	stats httpServerStats
	conf  ServerConfig

	startedAt time.Time

	mux http.Handler

	sync.Mutex
	sync.WaitGroup
}

func NewMockServer(c ServerConfig) *MockServer {
	var (
		s      MockServer
		mux    = mux.NewRouter()
		record = c.Record
	)

	s.conf = c
	s.stats.Name = c.Name

	for _, p := range c.Paths {
		var (
			path   = p.Path
			router = mux.Path(p.Path).Subrouter()
		)

		if len(p.Record) != 0 {
			record = p.Record
		}

		for _, mthd := range p.Methods {
			var (
				method  = mthd.Method
				methods = allMethods
			)

			if method != "" {
				methods = []string{method}
			}

			if len(c.Record) != 0 {
				record = c.Record
			}

			var (
				rs      = newResponseSequence(mthd.ReturnSequence)
				handler = s.newHandler(path, record, rs)
				route   = router.Methods(methods...)
			)

			s.Add(1)
			go func(rs *responseSequence) {
				<-rs.doneChan
				fmt.Println(1)
				s.Lock()
				defer s.Unlock()
				s.Done()
			}(rs)
			route.Handler(handler)
		}
	}

	fmt.Printf("POOP: %+v\n", *mux)

	s.mux = mux
	s.startedAt = time.Now()
	return &s
}

func (s *MockServer) newHandler(path string, record []string, rs *responseSequence) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.conf.Serial {
			s.Lock()
			defer s.Unlock()
		}

		var response = rs.next()
		if response == nil {
			zeroHandler(w, r)
			return
		}

		s.stats.recordRequest(path, r.Method, r, record)
		s.stats.inc(r.URL.Path, r.Method)

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
	})
}

func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *MockServer) ServerStats() interface{} {
	var ss = s.stats.ServerStats()
	ss.SampledAt = time.Now()
	ss.Duration = ss.SampledAt.Sub(s.startedAt).Seconds()

	return ss
}

func zeroHandler(w http.ResponseWriter, r *http.Request) {
	if hj, ok := w.(http.Hijacker); ok {
		conn, _, err := hj.Hijack()
		if err == nil {
			conn.Close()
			return
		}
	}
	w.WriteHeader(http.StatusGone)
}
