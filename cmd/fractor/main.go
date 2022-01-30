package main

import (
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"
)

type param struct {
	name      string
	shorthand string
	value     interface{}
	usage     string
	required  bool
}

var (
	Version    string
	Build      string
	rootParams = []param{
		{name: "config", shorthand: "c", value: ".", usage: "path to configuration file"},
		{name: "json-log", shorthand: "", value: false, usage: "output logs in json format"},
		{name: "verbose", shorthand: "", value: false, usage: "enable verbose logs"},
	}
)

var fractorVersion = &cobra.Command{
	Use:   "version",
	Short: "Print factor version and build sha",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🐾 version: %s build: %s \n", Version, Build)
	},
}

var fractorStart = &cobra.Command{
	Use:   "start",
	Short: "Start fractor device plugin",
	Run: func(cmd *cobra.Command, args []string) {

		if ret := nvml.Init(); ret != nvml.SUCCESS {
			log.Error(nvml.ErrorString(ret))
		}

		count, ret := nvml.DeviceGetCount()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
		}

		for i := 0; i < count; i++ {
			device, ret := nvml.DeviceGetHandleByIndex(i)
			if ret != nvml.SUCCESS {
				log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
			}

			uuid, ret := device.GetUUID()
			if ret != nvml.SUCCESS {
				log.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
			}

			log.Info("%v", uuid)
		}

		f := pkg.NewMetaFractorDevicePlugin()

		if err := f.Serve(); err != nil {
			log.Fatal(err)
		}

		if err := f.Register(); err != nil {
			log.Fatal(err)
		}

		for {
			time.Sleep(5 * time.Second)
		}

	},
}

var rootCmd = &cobra.Command{
	Use:   "fractor",
	Short: "Fractional accelerator device plugin",
}

func setParams(params []param, command *cobra.Command) {
	for _, param := range params {
		switch v := param.value.(type) {
		case int:
			command.PersistentFlags().IntP(param.name, param.shorthand, v, param.usage)
		case string:
			command.PersistentFlags().StringP(param.name, param.shorthand, v, param.usage)
		case bool:
			command.PersistentFlags().BoolP(param.name, param.shorthand, v, param.usage)
		}
		if err := viper.BindPFlag(param.name, command.PersistentFlags().Lookup(param.name)); err != nil {
			panic(err)
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	setParams(rootParams, rootCmd)
	rootCmd.AddCommand(fractorVersion)
	rootCmd.AddCommand(fractorStart)

}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("FRACTOR")
	setupLogging()
}

func setupLogging() {

	// Set log verbosity
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Set log format
	if viper.GetBool("json-log") {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetReportCaller(true)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				fileName := fmt.Sprintf(" [%s]", path.Base(frame.Function)+":"+strconv.Itoa(frame.Line))
				return "", fileName
			},
		})
	}

	// Logs are always goes to STDOUT
	log.SetOutput(os.Stdout)
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
