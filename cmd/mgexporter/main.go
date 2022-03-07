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
	rootParams = []param{}
	rootCmd    = &cobra.Command{
		Use:   "mgexporter",
		Short: "mgexporter - Metagpu metrics exporter",
	}
	version = &cobra.Command{
		Use:   "version",
		Short: "Print metagpu metric exporter version and build sha",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("🐾 version: %s build: %s \n", Version, Build)
		},
	}
	startParams = []param{
		{name: "metrics-addr", shorthand: "a", value: "0.0.0.0:2112", usage: "listen address"},
		{name: "mgsrv", shorthand: "s", value: "127.0.0.1:50052", usage: "metagpu device plugin gRPC server address"},
		{name: "token", shorthand: "t", value: "", usage: "metagpu server authenticate token"},
	}
	start = &cobra.Command{
		Use:   "start",
		Short: "start metagpu metrics exporter",
		Run: func(cmd *cobra.Command, args []string) {
			startExporter()
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	setParams(startParams, start)
	rootCmd.AddCommand(version)
	rootCmd.AddCommand(start)
	setParams(rootParams, rootCmd)
}

func initConfig() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MG_EXPORTER")
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
