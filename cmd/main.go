package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/cheddartv/mockarena/internal/server"
	mhttp "github.com/cheddartv/mockarena/internal/server/http"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s FILE", os.Args[0])
	}

	conf, err := server.ParseConfigFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for _, mockConf := range conf.Mocks {
		switch c := mockConf.Mock.(type) {
		case *mhttp.ServerConfig:
			wg.Add(1)
			go func() {
				var s = mhttp.NewMockServer(*c)
				http.ListenAndServe(fmt.Sprintf("%s:%d", "", c.Port), s)
				wg.Done()
			}()
		default:
			panic("unimplemented")
		}
	}

	wg.Wait()
}
