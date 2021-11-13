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
Usage: mockarena [-h] [-c <path>] [-f <output_template>]

Options:
	-h --help  			Show this help message.
	-c --conf  			Path to the mockarena file describing the services to mock.
	-f --template			Template the output.	 

Templating:
The template passed to the -f option operates on

%s
Where objects in the Mocks array are of type

%s
Templating Functions:
%s
`

var (
	filePath string
	tmplt    string
	help     bool
)

func init() {
	flag.StringVarP(&filePath, "conf", "c", "./mockarena.yaml", "Path to the mockarena file.")
	flag.StringVarP(&tmplt, "template", "f", "", "Template the output.")
	flag.BoolVarP(&help, "help", "h", false, "Show help message.")
	flag.Parse()
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

	var output = output{
		Mocks:    make([]interface{}, len(mocks)),
		Duration: time.Since(startTime).Seconds(),
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
		var encoder = json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "\t")
		if err = encoder.Encode(output); err != nil {
			panic(err)
		}
	}

	retCode = 0
}

type output struct {
	Mocks    []interface{}
	Duration float64
}

func showHelp() {
	fmt.Println(
		fmt.Sprintf(helpMsg,
			ttools.GenerateUsageUndecorated(output{}),
			ttools.GenerateUsageUndecorated(mhttp.ServerStats{}),
			ttools.TemplateFunctionHelp(ttools.NewConfig(ttools.OptAllFns)),
		),
	)
}
