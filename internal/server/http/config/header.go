package config

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
