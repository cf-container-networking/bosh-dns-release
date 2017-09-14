package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"bosh-dns/healthcheck/healthexecutable"
	"bosh-dns/healthcheck/healthserver"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

var healthServer healthserver.HealthServer

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	const logTag = "healthcheck"

	logger := boshlog.NewAsyncWriterLogger(boshlog.LevelDebug, os.Stdout)
	defer logger.FlushTimeout(5 * time.Second)
	logger.Info(logTag, "Initializing")

	config, err := getConfig()
	if err != nil {
		logger.Error(logTag, fmt.Sprintf("Error: %v", err.Error()))
		return 1
	}

	fs := boshsys.NewOsFileSystem(logger)
	healthExecutable := healthexecutable.NewHealthExecutable()
	healthServer = healthserver.NewHealthServer(logger, fs, config.HealthFileName, healthExecutable)

	healthServer.Serve(config)
	return 0
}

func getConfig() (*healthserver.HealthCheckConfig, error) {
	var configFile string
	var config *healthserver.HealthCheckConfig

	if len(os.Args) > 1 {
		configFile = os.Args[1]
	} else {
		return nil, errors.New("Expected config file path argument")
	}

	configRaw, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Couldn't open config file for health. error: %s", err)
	}

	config = &healthserver.HealthCheckConfig{}
	err = json.Unmarshal(configRaw, config)
	if err != nil {
		return nil, fmt.Errorf("Couldn't decode config file for health. error: %s", err)
	}

	return config, nil
}
