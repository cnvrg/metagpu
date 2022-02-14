package main

import (
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configCmdParams = []param{
		{name: "metagpu", shorthand: "m", value: 0, usage: "set metagpus quantity (gpu shares)"},
	}
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "change configs on running metagpu device plugin instance",
	Run: func(cmd *cobra.Command, args []string) {
		patchConfigs()
	},
}

func patchConfigs() {
	if viper.GetInt32("metagpu") != 0 {
		metaGpus := viper.GetInt32("metagpu")
		log.Info(metaGpus)
		conn, err := GetGrpcMetaGpuSrvClientConn()
		if err != nil {
			log.Fatalf("can't initiate connection to metagpu server, %s", err)
		}
		device := pbdevice.NewDeviceServiceClient(conn)

		request := &pbdevice.PatchConfigsRequest{MetaGpus: metaGpus}
		if _, err := device.PatchConfigs(authenticatedContext(), request); err != nil {
			log.Error(err)
		}
	}
}
