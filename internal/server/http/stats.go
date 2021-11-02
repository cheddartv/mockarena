package http

import "sync"

type httpServerStats struct {
	Count      uint
	PathCounts map[string]uint

	sync.Mutex
}

func (ss *httpServerStats) inc(path string) {
	ss.Lock()
	ss.Count++
	if ss.PathCounts == nil {
		ss.PathCounts = make(map[string]uint)
	}
	ss.PathCounts[path]++
	ss.Unlock()
}
