package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type param struct {
	name      string
	shorthand string
	value     interface{}
	usage     string
	required  bool
}

const (
	outJSON  = "json"
	outTable = "table"
	outRaw   = "raw"
)

const (
	// flagAddr defines output format.
	flagOutput = "output"
	// flagOutputS short form of flagOutput.
	flagOutputS = "o"
	// flagJSONLog enables log json.
	flagJSONLog = "json-log"
	// flagVerbose enables verbose logging.
	flagVerbose = "verbose"
	// flagAddr MetaGPU server address, port.
	flagAddr = "addr"
	// flagAddrS short form of flagAddr.
	flagAddrS = "s"
	// flagToken authentication token.
	flagToken = "token"
	// flagTokenS short form of flagToken.
	flagTokenS = "t"
	// flagPrettyOut enables indented JSON output for humans.
	flagPrettyOut = "pretty"
)

var (
	Version    string
	Build      string
	rootParams = []param{
		{name: flagJSONLog, shorthand: "", value: false, usage: "output logs in json format"},
		{name: flagVerbose, shorthand: "", value: false, usage: "enable verbose logs"},
		{name: flagAddr, shorthand: flagAddrS, value: "localhost:50052", usage: "address to access the metagpu server"},
		{name: flagToken, shorthand: flagTokenS, value: "", usage: "authentication token"},
		{name: flagOutput, shorthand: flagOutputS, value: outTable, usage: "output format, one of: table|json|raw"},
		{name: flagPrettyOut, shorthand: "", value: false, usage: "pretty output for JSON"},
	}
)

var metaGpuCtlVersion = &cobra.Command{
	Use:   "version",
	Short: "Print metagpuctl version and build sha",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🐾 version: %s build: %s \n", Version, Build)
	},
}

var rootCmd = &cobra.Command{
	Use:   "mgctl",
	Short: "mgctl - cli client for metagpu management and monitoring",
}

func init() {
	cobra.OnInitialize(initConfig)
	setParams(configCmdParams, configCmd)
	setParams(processGetParams, processesGetCmd)
	setParams(rootParams, rootCmd)
	// processes
	getCmd.AddCommand(processesGetCmd)
	getCmd.AddCommand(getDevicesCmd)
	// root commands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(enforceCmd)
	rootCmd.AddCommand(killCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(metaGpuCtlVersion)

}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MG_CTL")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	setupLogging()
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

func setupLogging() {

	// Set log verbosity
	if viper.GetBool(flagVerbose) {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
				fileName := fmt.Sprintf(" [%s]", path.Base(frame.Function)+":"+strconv.Itoa(frame.Line))
				return "", fileName
			},
		})
	} else {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	}

	// Set log format
	if viper.GetBool(flagJSONLog) {
		log.SetFormatter(&log.JSONFormatter{})
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
