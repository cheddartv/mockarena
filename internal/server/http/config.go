package http

type ServerConfig struct {
	Name           string      `yaml:"name"`
	Port           int         `yaml:"port"`
	Serial         bool        `yaml:"serial"`
	ReturnSequence []*Response `yaml:"returnSequence"`
}

func NewServerConfig(m map[string]interface{}) (*ServerConfig, error) {
	var conf ServerConfig
	if x, ok := m["name"]; ok {
		conf.Name = x.(string)
	}
	if x, ok := m["port"]; ok {
		conf.Port = x.(int)
	}
	if x, ok := m["serial"]; ok {
		conf.Serial = x.(bool)
	}

	if x, ok := m["returnSequence"]; ok {
		var responses []*Response

		for _, iface := range x.([]interface{}) {
			response, err := newResponse(iface.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}

			responses = append(responses, response)
		}

		conf.ReturnSequence = responses
	}

	return &conf, nil
}

func (sc *ServerConfig) Type() string {
	return "HTTP"
}
