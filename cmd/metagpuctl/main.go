package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
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
		{name: "json-log", shorthand: "", value: false, usage: "output logs in json format"},
		{name: "verbose", shorthand: "", value: false, usage: "enable verbose logs"},
		{name: "metagpu-server-addr", shorthand: "s", value: "localhost:50052", usage: "address to access the metagpu server"},
		{name: "token", shorthand: "t", value: "", usage: "authentication token"},
		{name: "output", shorthand: "o", value: "table", usage: "output format, one of: table|json|raw"},
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
	setParams(rootParams, rootCmd)
	// processes
	ListCmd.AddCommand(ProcessesListCmd)
	// root commands
	rootCmd.AddCommand(ListCmd)
	rootCmd.AddCommand(PingCmd)
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
	if viper.GetBool("verbose") {
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
	if viper.GetBool("json-log") {
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
