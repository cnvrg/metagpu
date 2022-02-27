package main

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping server to check connectivity",
	Run: func(cmd *cobra.Command, args []string) {
		conn := utils.GetGrpcMetaGpuSrvClientConn(viper.GetString("addr"))
		if conn == nil {
			log.Fatalf("can't initiate connection to metagpu server")
		}
		defer conn.Close()
	},
}
