package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-cpi-go/apiv1"
	"github.com/cloudfoundry/bosh-cpi-go/rpc"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"

	"bosh-vmrun-cpi/action"
	"bosh-vmrun-cpi/config"
	"bosh-vmrun-cpi/driver"
	"bosh-vmrun-cpi/stemcell"
	"bosh-vmrun-cpi/vm"
	"bosh-vmrun-cpi/vmx"
)

var (
	configPathOpt       = flag.String("configPath", "", "Path to configuration file")
	configBase64JsonOpt = flag.String("configBase64JSON", "", "Base64-encoded JSON string of configuration")
	versionOpt          = flag.Bool("version", false, "Version")

	//set by X build flag
	version string
)

func main() {
	var err error

	rand.Seed(time.Now().UTC().UnixNano()) // todo MAC generation

	logLevel, err := boshlog.Levelify(os.Getenv("BOSH_LOG_LEVEL"))
	if err != nil {
		logLevel = boshlog.LevelDebug
	}

	logger := boshlog.NewWriterLogger(logLevel, os.Stderr)
	defer logger.HandlePanic("Main")

	flag.Parse()

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	fs := boshsys.NewOsFileSystem(logger)
	cmdRunner := boshsys.NewExecCmdRunner(logger)
	compressor := boshcmd.NewTarballCompressor(cmdRunner, fs)
	uuidGen := boshuuid.NewGenerator()

	var configJson string
	if *configPathOpt != "" {
		configJsonBytes, err := ioutil.ReadFile(*configPathOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "loading cfg", err)
			os.Exit(1)
		}
		configJson = string(configJsonBytes)
	} else if *configBase64JsonOpt != "" {
		configJsonBytes, err := base64.StdEncoding.DecodeString(*configBase64JsonOpt)
		if err != nil {
			logger.ErrorWithDetails("main", "base64 decoding cfg", err)
			os.Exit(1)
		}
		configJson = string(configJsonBytes)
	} else {
		logger.Error("main", "config option required")
		os.Exit(1)
	}

	cpiConfig, err := config.NewConfigFromJson(configJson)
	if err != nil {
		logger.ErrorWithDetails("main", "config JSON is invalid", err)
		os.Exit(1)
	}

	driverConfig := driver.NewConfig(cpiConfig)
	stemcellConfig := stemcell.NewConfig(cpiConfig)
	retryFileLock := driver.NewRetryFileLock(logger)

	ovftoolRunner := driver.NewOvftoolRunner(driverConfig.OvftoolPath(), cmdRunner, logger)
	if err = ovftoolRunner.Configure(); err != nil {
		logger.ErrorWithDetails("main", "ovftool is invalid", err)
		os.Exit(1)
	}

	vmrunRunner := driver.NewVmrunRunner(driverConfig.VmrunPath(), retryFileLock, logger)
	if err = vmrunRunner.Configure(); err != nil {
		logger.ErrorWithDetails("main", "vmrun is invalid", err)
		os.Exit(1)
	}

	var cloneRunner driver.CloneRunner
	if vmrunRunner.IsPlayer() {
		cloneRunner = ovftoolRunner
	} else {
		cloneRunner = vmrunRunner
	}

	vmxBuilder := vmx.NewVmxBuilder(logger)
	driverClient := driver.NewClient(vmrunRunner, ovftoolRunner, cloneRunner, vmxBuilder, driverConfig, logger)
	stemcellClient := stemcell.NewClient(compressor, fs, logger)
	stemcellStore := stemcell.NewStemcellStore(stemcellConfig, compressor, fs, logger)
	agentEnvFactory := apiv1.NewAgentEnvFactory()
	agentSettings := vm.NewAgentSettings(fs, logger, agentEnvFactory)
	cpiFactory := action.NewFactory(driverClient, stemcellClient, stemcellStore, agentSettings, agentEnvFactory, cpiConfig, fs, uuidGen, logger)

	cli := rpc.NewFactory(logger).NewCLI(cpiFactory)

	err = cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}
