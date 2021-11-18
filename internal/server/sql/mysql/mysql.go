package mysql

import (
	"errors"
	"strings"
	"time"
)

type Response interface {
	_isResponse()
}

type _response struct{}

func (*_response) _isResponse() {}

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
