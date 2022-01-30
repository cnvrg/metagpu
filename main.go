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

	},
}



var rootCmd = &cobra.Command{
	Use:   "fractor",
	Short: "Fractional accelerator device plugin",
}

func initConfig() {

	viper.AutomaticEnv()
	//viper.SetConfigName("config")
	//viper.SetConfigType("yaml")
	//viper.AddConfigPath("./config")
	//viper.AddConfigPath(viper.GetString("config"))
	viper.SetEnvPrefix("FRACTOR_DP")
	//viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setupLogging()
	//err := viper.ReadInConfig()
	//if err != nil {
	//	log.Fatalf("config file not found, err: %s", err)
	//}
	//viper.WatchConfig()
	//viper.OnConfigChange(func(e fsnotify.Event) {
	//	log.Infof("config file changed: %s, emit reconcile event for all clusters and tenants", e.Name)
	//})

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
