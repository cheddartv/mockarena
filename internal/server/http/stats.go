package http

import (
	"io"
	"net/http"
	"sync"
	"time"
)

func pullStatsFromRequest(r *http.Request, stats []string) map[string]interface{} {
	var m = make(map[string]interface{})

	for _, stat := range stats {
		switch stat {
		case "method":
			m["Method"] = r.Method
		case "url":
			m["URL"] = r.URL
		case "proto":
			m["Proto"] = r.Proto
		case "header":
			m["Header"] = r.Header
		case "body":
			var body, _ = io.ReadAll(r.Body)
			r.Body.Close()
			m["Body"] = string(body)
		case "contentlength":
			m["ContentLength"] = r.ContentLength
		case "host":
			m["Host"] = r.Host
		case "form":
			r.ParseForm()
			m["Form"] = r.Form
		case "postform":
			r.ParseForm()
			m["PostForm"] = r.PostForm
		case "trailer":
			m["Trailer"] = r.Trailer
		case "remoteaddr":
			m["RemoteAddr"] = r.RemoteAddr
		case "requesturi":
			m["RequestURI"] = r.RequestURI
		case "tls":
			m["TLS"] = r.TLS
		case "time":
			m["Time"] = time.Now()
		default:
			continue
		}
	}

	return m
}

type MethodStats struct {
	Name     string
	Count    uint
	Requests []map[string]interface{}
}

type PathStats struct {
	Name         string
	Count        uint
	MethodsStats map[string]*MethodStats
}

type httpServerStats struct {
	Name       string
	Count      uint
	PathsStats map[string]*PathStats

	sync.Mutex
}

func (ss *httpServerStats) inc(path, method string) {
	ss.Lock()
	defer ss.Unlock()

	ss.Count++
	if ss.PathsStats == nil {
		ss.PathsStats = make(map[string]*PathStats)
	}

	ps, ok := ss.PathsStats[path]
	if !ok {
		ps = &PathStats{
			Name:         path,
			MethodsStats: make(map[string]*MethodStats),
		}
		ss.PathsStats[path] = ps
	}
	ps.Count++

	methodsStats, ok := ps.MethodsStats[method]
	if !ok {
		methodsStats = &MethodStats{
			Name:     method,
			Requests: make([]map[string]interface{}, 0),
		}

		ps.MethodsStats[method] = methodsStats
	}

	ps.MethodsStats[method].Count++
}

func (ss *httpServerStats) recordRequest(path, method string, r *http.Request, fields []string) {
	ss.Lock()
	defer ss.Unlock()

	if ss.PathsStats == nil {
		ss.PathsStats = make(map[string]*PathStats)
	}

	pathsStats, ok := ss.PathsStats[path]
	if !ok {
		pathsStats = &PathStats{
			Name:         path,
			MethodsStats: make(map[string]*MethodStats),
		}

		ss.PathsStats[path] = pathsStats
	}

	methodsStats, ok := pathsStats.MethodsStats[method]
	if !ok {
		methodsStats = &MethodStats{
			Name:     method,
			Requests: make([]map[string]interface{}, 0),
		}

		pathsStats.MethodsStats[method] = methodsStats
	}

	var rs = pullStatsFromRequest(r, fields)
	methodsStats.Requests = append(methodsStats.Requests, rs)
}

func (ss *httpServerStats) ServerStats() *ServerStats {
	return &ServerStats{
		Name:       ss.Name,
		Count:      ss.Count,
		PathsStats: ss.PathsStats,
	}
}

type ServerStats struct {
	Name       string
	Count      uint
	Duration   float64
	SampledAt  time.Time
	PathsStats map[string]*PathStats
}
