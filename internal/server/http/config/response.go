package config

import (
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
