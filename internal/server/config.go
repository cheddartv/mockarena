package server

import (
	"fmt"
	"os"

	httpconfig "github.com/cheddartv/mockarena/internal/server/http/config"
	"gopkg.in/yaml.v2"
)

type Configuration struct {
	ReportPath string `yaml:"reportPath"`
	Port       int    `yaml:"port"`
	Host       string `yaml:"host"`

	Mocks []MockConfiguration `yaml:"mocks"`
}

type MockConfiguration struct {
	Mock interface {
		Type() string
	}
}

func (mc *MockConfiguration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var m map[string]interface{}
	if err := unmarshal(&m); err != nil {
		return err
	}

	switch t := m["type"]; t {
	case "HTTP":
		conf, err := httpconfig.NewServerConfig(m)
		if err != nil {
			return err
		}

		mc.Mock = conf
	default:
		return fmt.Errorf("unsupported mock type: %s", t)
	}

	return nil
}

func ParseConfigFile(path string) (*Configuration, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var c Configuration
	if err := yaml.NewDecoder(file).Decode(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
