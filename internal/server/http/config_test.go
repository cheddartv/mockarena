package http

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestNewServerConfig(t *testing.T) {
	var tests = []struct {
		name           string
		imap           map[string]interface{}
		expectedConfig ServerConfig
		err            error
	}{
		{
			name: "full",
			imap: map[string]interface{}{
				"name": "TEST_NAME",
				"port": 27182,
				"returnSequence": []interface{}{
					map[interface{}]interface{}{
						"header": []interface{}{
							map[interface{}]interface{}{
								"key":   "TEST_KEY",
								"value": "TEST_VALUE",
							},
						},
						"repeat": map[interface{}]interface{}{
							"count": 3,
							"for":   "1m20s",
							"until": "2021-11-01 02:03:04",
						},
						"delay":  "12s",
						"body":   "TEST_BODY",
						"status": 200,
					},
				},
			},
			expectedConfig: ServerConfig{
				Name: "TEST_NAME",
				Port: 27182,
				Paths: []*Path{
					{
						Path: "",
						ReturnSequence: []*Response{
							{
								Header: []Header{
									{
										Key:   "TEST_KEY",
										Value: "TEST_VALUE",
									},
								},
								Repeat: Repeat{
									Count: 3,
									Until: time.Date(2021, time.November, 1, 2, 3, 4, 0, time.UTC),
									For:   80 * time.Second,
								},
								Delay:  12 * time.Second,
								Body:   []byte("TEST_BODY"),
								Status: 200,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				is             = is.New(t)
				imap           = test.imap
				expectedConfig = test.expectedConfig
			)

			sc, err := NewServerConfig(imap)
			is.NoErr(err)

			is.Equal(sc.Name, expectedConfig.Name)
			is.Equal(sc.Port, expectedConfig.Port)

			var path = sc.Paths[0]
			for idx, r := range path.ReturnSequence {
				is.Equal(*r, *path.ReturnSequence[idx])
			}
		})
	}
}
