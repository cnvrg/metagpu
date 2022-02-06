package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
	}
)

var metaGpuControllerVersion = &cobra.Command{
	Use:   "version",
	Short: "Print metagpu-controller version and build sha",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🐾 version: %s build: %s \n", Version, Build)
	},
}

var metaGpuControllerStart = &cobra.Command{
	Use:   "start",
	Short: "Start metagpu controller",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var rootCmd = &cobra.Command{
	Use:   "metagpu",
	Short: "Metagpu - fractional accelerator device plugin",
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
