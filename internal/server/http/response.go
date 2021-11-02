package http

import (
	"sync"
	"time"
)

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func NewHeader(imap map[interface{}]interface{}) (*Header, error) {
	var header Header

	header.Key = imap["key"].(string)
	header.Value = imap["value"].(string)

	return &header, nil
}

type Repeat struct {
	Count uint          `yaml:"count"`
	Until time.Time     `yaml:"until"`
	For   time.Duration `yaml:"for"`
}

func NewRepeat(imap map[interface{}]interface{}) (*Repeat, error) {
	var repeat Repeat

	if x, ok := imap["until"]; ok {
		t, err := time.Parse("2006-01-02 15:04:05", x.(string))
		if err != nil {
			return nil, err
		}

		repeat.Until = t
	}

	if x, ok := imap["for"]; ok {
		d, err := time.ParseDuration(x.(string))
		if err != nil {
			return nil, err
		}

		repeat.For = d
	}

	if x, ok := imap["count"]; ok {
		repeat.Count = uint(x.(int))
	}

	return &repeat, nil
}

type Response struct {
	Header []Header      `yaml:"header"`
	Body   []byte        `yaml:"body"`
	Delay  time.Duration `yaml:"delay"`
	Status int           `yaml:"status"`

	Repeat Repeat `yaml:"repeat"`
}

func NewResponse(imap map[interface{}]interface{}) (*Response, error) {
	var response Response

	if x, ok := imap["header"]; ok {
		var (
			ifaces  = x.([]interface{})
			headers = make([]Header, len(ifaces))
		)

		for idx, iface := range x.([]interface{}) {
			header, err := NewHeader(iface.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}

			headers[idx] = *header
		}

		response.Header = headers
	}

	if x, ok := imap["repeat"]; ok {
		repeat, err := NewRepeat(x.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}

		response.Repeat = *repeat
	} else {
		response.Repeat.Count = 1
	}

	if x, ok := imap["delay"]; ok {
		d, err := time.ParseDuration(x.(string))
		if err != nil {
			return nil, err
		}

		response.Delay = d
	}

	if x, ok := imap["body"]; ok {
		response.Body = []byte(x.(string))
	}

	if x, ok := imap["status"]; ok {
		response.Status = x.(int)
	}

	return &response, nil
}

type ResponseSequence struct {
	Responses []*Response

	responseCounts       map[*Response]uint
	currentResponseStart *time.Time

	sync.RWMutex
}

func (rs *ResponseSequence) next() *Response {
	rs.Lock()
	defer rs.Unlock()

	return rs._next()
}

func (rs *ResponseSequence) _next() *Response {
	if len(rs.Responses) == 0 {
		return nil
	}

	var (
		response = rs.Responses[0]
		repeat   = response.Repeat
	)

	switch {
	case 0 < repeat.Count:
		if rs.responseCounts == nil {
			rs.responseCounts = make(map[*Response]uint)
		}

		if repeat.Count <= rs.responseCounts[response] {
			// response count limit exceeded, discard and move on
			rs.Responses = rs.Responses[1:]
			return rs._next()
		}

		rs.responseCounts[response]++

		return response
	case !repeat.Until.IsZero():
		var cutOff = repeat.Until

		if !time.Now().Before(cutOff) {
			// time limit exceeded, discard and move on
			rs.Responses = rs.Responses[1:]
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

			rs.Responses = rs.Responses[1:]
			return rs._next()
		}

		return response
	}

	return nil
}
