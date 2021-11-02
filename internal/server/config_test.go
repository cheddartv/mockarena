package server

import (
	"testing"

	"github.com/matryer/is"
	"gopkg.in/yaml.v2"

	httpconfig "github.com/cheddartv/mockarena/internal/server/http/config"
)

func TestConfiguration_UnmarshalYAML(t *testing.T) {
	var (
		yamlPayload = `
reportPath: "./out.json"
port: 8080
mocks:
   - name: "web_server"
     port: 4000
     type: "HTTP"
     record: ["address", "header", "time", "body"]
     returnSequence:
        - &default_http
          header: &default_header
            - key: "Content-Type"
              value: "text/plain"
          body: "OK"
          delay: 500ms
          status: 200
          repeat:
            until: "2021-11-01 16:55:00"
            for: 30s
            count: 2
        - header: *default_header
          body: "ERROR"
          status: 500
        - <<: *default_http
          repeat:
            for: 10s
`
		is = is.New(t)

		c Configuration
	)

	err := yaml.Unmarshal([]byte(yamlPayload), &c)
	is.NoErr(err)

	is.Equal(c.ReportPath, "./out.json")
	is.Equal(c.Port, 8080)
	is.Equal(len(c.Mocks), 1)
	_, ok := c.Mocks[0].Mock.(*httpconfig.ServerConfig)
	is.True(ok)
}
