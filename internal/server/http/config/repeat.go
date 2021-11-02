package config

import "time"

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
