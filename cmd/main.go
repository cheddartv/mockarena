package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cheddartv/mockarena/internal/server"
	mhttp "github.com/cheddartv/mockarena/internal/server/http"
	ttools "github.com/intel/tfortools"
	flag "github.com/spf13/pflag"
)

const helpMsg = `
Usage: mockarena [-h] [-f <path>] [output_template]
	output_template: A Go template string used to format output.

Options:
	-h --help  Show this help message.
	-f --file  Path to the mockarena file describing the services to mock.
`

var (
	filePath string
	tmplt    string
	help     bool
)

func init() {
	flag.StringVarP(&filePath, "file", "f", "./mockarena.yaml", "Path to the mockarena file.")
	flag.BoolVarP(&help, "help", "h", false, "Show help message.")
	flag.Parse()
	tmplt = flag.Arg(0)
}

func main() {
	var retCode = 1
	defer func() {
		os.Exit(retCode)
	}()

	if help {
		showHelp()
		retCode = 0
		return
	}

	conf, err := server.ParseConfigFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	var (
		startTime = time.Now()
		wg        sync.WaitGroup
		mocks     = make([]server.Server, len(conf.Mocks))
	)

	for idx, mockConf := range conf.Mocks {
		switch c := mockConf.Mock.(type) {
		case *mhttp.ServerConfig:
			wg.Add(1)
			var s = mhttp.NewMockServer(*c)
			mocks[idx] = s

			go func(s *mhttp.MockServer) {
				go http.ListenAndServe(fmt.Sprintf("%s:%d", "", c.Port), s)
				s.Wait()
				fmt.Println(10)
				wg.Done()
			}(s)
		default:
			panic("unimplemented")
		}
	}

	var (
		sigChan  = make(chan os.Signal, 1)
		doneChan = make(chan struct{})
	)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT)
	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case <-sigChan:
	case <-doneChan:
	}

	var output = struct {
		Mocks    []interface{}
		Duration float64
	}{
		Mocks:    make([]interface{}, len(mocks)),
		Duration: time.Now().Sub(startTime).Seconds(),
	}
	for idx := range output.Mocks {
		output.Mocks[idx] = mocks[idx].ServerStats()
	}

	switch tmplt {
	default:
		err := ttools.OutputToTemplate(
			os.Stdout,
			"stats",
			tmplt,
			output,
			ttools.NewConfig(ttools.OptAllFns),
		)
		os.Stdout.Write([]byte("\n"))

		if err != nil {
			log.Fatal(err)
		}
	case "":
		var data, err = json.MarshalIndent(output, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(data))
	}

	retCode = 0
}

func showHelp() {
	fmt.Println(helpMsg)
}
