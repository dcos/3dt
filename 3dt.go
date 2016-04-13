package main

import (

	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/mesosphere/3dt/api"
	"net/http"
)

const	Version string = "0.0.9"

func GetVersion() string {
	return (fmt.Sprintf("Version: %s", Version))
}

func RunDiag(config api.Config) int {
	var exitCode int = 0
	units, err := api.GetUnitsProperties(&config)
	if err != nil {
		log.Error(err)
		return 1
	}
	for _, unit := range units.Array {
		if unit.UnitHealth != 0 {
			fmt.Printf("[%s]: %s %s\n", unit.UnitId, unit.UnitTitle, unit.UnitOutput)
			exitCode = 1
		}
	}
	return exitCode
}

func main() {
	// a message channel to ensure we can start pulling safely
	readyChan := make(chan bool, 1)

	// load config with default values
	config, err := api.LoadDefaultConfig(os.Args, Version)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	// print version and exit
	if config.FlagVersion {
		fmt.Println(GetVersion())
		os.Exit(0)
	}

	// run local diagnostics, verify all systemd units are healthy.
	if config.FlagDiag {
		os.Exit(RunDiag(config))
	}

	// set verbose (debug) output.
	if config.FlagVerbose {
		log.SetLevel(log.DebugLevel)
	}

	// start pulling every 60 seconds.
	if config.FlagPull {
		puller := api.PullType{}
		go api.StartPullWithInterval(config, &puller, readyChan)
	}

	// start diagnostic server and expose endpoints.
	log.Info("Start 3DT")
	go api.StartUpdateHealthReport(config, readyChan)
	router := api.NewRouter(&config)
	log.Infof("Exposing 3DT API on 0.0.0.0:%d", config.FlagPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.FlagPort), router))
}
