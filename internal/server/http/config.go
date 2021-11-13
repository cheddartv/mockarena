package http

type Method struct {
	Method         string      `yaml:method"`
	Record         []string    `yaml:"record"`
	ReturnSequence []*Response `yaml:"returnSequence"`
}

func newMethod(m map[interface{}]interface{}) (*Method, error) {
	var mthd Method

	if x, ok := m["method"]; ok {
		mthd.Method = x.(string)
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

		mthd.ReturnSequence = responses
	}

	if x, ok := m["record"]; ok {
		var ifaces = x.([]interface{})

		mthd.Record = make([]string, len(ifaces))

		for idx := range ifaces {
			mthd.Record[idx] = ifaces[idx].(string)
		}
	}

	return &mthd, nil
}

type Path struct {
	Path    string    `yaml:"path"`
	Record  []string  `yaml:"record"`
	Methods []*Method `yaml:"methods"`
}

func newPath(m map[interface{}]interface{}) (*Path, error) {
	var p Path

	if x, ok := m["path"]; ok {
		p.Path = x.(string)
	}

	if p.Path == "" {
		p.Path = "/"
	}

	if x, ok := m["methods"]; ok {
		var methods []*Method

		for _, iface := range x.([]interface{}) {
			method, err := newMethod(iface.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}

			methods = append(methods, method)
		}

		p.Methods = methods
	}

	if x, ok := m["record"]; ok {
		var ifaces = x.([]interface{})

		p.Record = make([]string, len(ifaces))

		for idx := range ifaces {
			p.Record[idx] = ifaces[idx].(string)
		}
	}

	return &p, nil
}

type ServerConfig struct {
	Name   string   `yaml:"name"`
	Record []string `yaml:"record"`
	Port   int      `yaml:"port"`
	Serial bool     `yaml:"serial"`
	Paths  []*Path  `yaml:"paths"`
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

	if x, ok := m["record"]; ok {
		var ifaces = x.([]interface{})

		conf.Record = make([]string, len(ifaces))

		for idx := range ifaces {
			conf.Record[idx] = ifaces[idx].(string)
		}
	}

	if x, ok := m["paths"]; ok {
		var paths []*Path

		for _, iface := range x.([]interface{}) {
			path, err := newPath(iface.(map[interface{}]interface{}))
			if err != nil {
				return nil, err
			}

			paths = append(paths, path)
		}

		conf.Paths = paths
	}

	return &conf, nil
}

func (sc *ServerConfig) Type() string {
	return "HTTP"
}
