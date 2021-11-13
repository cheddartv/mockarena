package http

import (
	"errors"
	"strings"
	"time"
)

type Response struct {
	Header []Header      `yaml:"header"`
	Body   []byte        `yaml:"body"`
	Delay  time.Duration `yaml:"delay"`
	Status int           `yaml:"status"`
	Proxy  *Proxy        `yaml:"proxy"`
	Repeat Repeat        `yaml:"repeat"`
}

func newResponse(imap map[interface{}]interface{}) (*Response, error) {
	var response Response

	if x, ok := imap["header"]; ok {
		var (
			ifaces  = x.([]interface{})
			headers = make([]Header, len(ifaces))
		)

		for idx, iface := range x.([]interface{}) {
			header, err := newHeader(iface.(map[interface{}]interface{}))
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

	if x, ok := imap["proxy"]; ok {
		proxy, err := NewProxy(x.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}

		response.Proxy = proxy
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

type Repeat struct {
	Count   uint          `yaml:"count"`
	Until   time.Time     `yaml:"until"`
	For     time.Duration `yaml:"for"`
	Forever string        `yaml:"forever"`
}

func NewRepeat(imap map[interface{}]interface{}) (*Repeat, error) {
	var repeat Repeat

	if x, ok := imap["until"]; ok {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", x.(string), time.Local)
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

	if x, ok := imap["forever"]; ok {
		var s = x.(string)
		s = strings.ReplaceAll(s, "-", "")
		s = strings.ReplaceAll(s, "_", "")
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ToLower(s)

		if s != "nonblocking" && s != "blocking" {
			return nil, errors.New("repeat.forever must be either \"blocking\" or \"nonblocking\"")
		}

		repeat.Forever = s
	}

	return &repeat, nil
}

type Header struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func newHeader(imap map[interface{}]interface{}) (*Header, error) {
	var header Header

	header.Key = imap["key"].(string)
	header.Value = imap["value"].(string)

	return &header, nil
}
