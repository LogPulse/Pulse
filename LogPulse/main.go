package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/gophergala2016/Pulse/LogPulse/api"
	"github.com/gophergala2016/Pulse/LogPulse/config"
	"github.com/gophergala2016/Pulse/LogPulse/email"
	"github.com/gophergala2016/Pulse/LogPulse/file"
	"github.com/gophergala2016/Pulse/pulse"
)

var (
	runAPI      bool
	outputFile  string
	buffStrings []string
	logList     []string
)

func init() {
	flag.BoolVar(&runAPI, "api", false, "Turn on API mode")
	flag.Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(0)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("main.init: Could not load the config.\n %v", err))
	}

	logList = cfg.LogList
	outputFile = cfg.OutputFile
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
			os.Exit(0)
		}
	}()

	if len(flag.Args()) == 0 && !runAPI {
		if len(logList) == 0 {
			panic(fmt.Errorf("main.main: Must supply a list of log files in the config"))
		}
		startPulse(logList)
	} else if runAPI {
		startAPI()
	} else {
		startPulse(flag.Args())
	}
}

func startAPI() {
	api.Start()
}

func startPulse(filenames []string) {
	checkList(filenames)
	stdIn := make(chan string)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// On keyboard interrup cleanup the program
	go func() {
		for range c {
			fmt.Println("Exiting for Keyboard Interupt")
			os.Exit(0)
		}
	}()

	pulse.Run(stdIn, email.Send)
	for _, filename := range filenames {
		line := make(chan string)
		file.Read(filename, line)
		for l := range line {
			stdIn <- l
		}
	}
	close(stdIn)
}

func checkList(filenames []string) {
	for i, filename := range filenames {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			panic(fmt.Errorf("main.checkList: %s", err))
		}
		if len(filename) > 3 && filename[len(filename)-3:] == ".gz" {
			if err := file.UnGZip(filename); err != nil {
				panic(fmt.Errorf("main.checkList: %s", err))
			}
			if _, err := os.Stat(filename[:len(filename)-3]); os.IsNotExist(err) {
				panic(fmt.Errorf("main.checkList: %s", err))
			}
			filenames[i] = filename[:len(filename)-3]
		}
	}
}
