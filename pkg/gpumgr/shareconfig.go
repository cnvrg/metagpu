package gpumgr

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type DeviceSharingConfig struct {
	Uuid         []string
	ResourceName string
	MetaGpus     int
}

type DevicesSharingConfigs struct {
	Configs []*DeviceSharingConfig
}

func NewDeviceSharingConfig() *DevicesSharingConfigs {
	var cfg []*DeviceSharingConfig
	if err := viper.UnmarshalKey("deviceSharing", cfg); err != nil {
		log.Fatal(err)
	}
	return &DevicesSharingConfigs{Configs: cfg}
}

func (c *DevicesSharingConfigs) getDeviceSharingConfigs(devUuid string) (*DeviceSharingConfig, error) {
	// TODO: add support for wildcard
	for _, devCfg := range c.Configs {
		for _, uuid := range devCfg.Uuid {
			if uuid == devUuid {
				return devCfg, nil
			}
		}
	}
	return nil, fmt.Errorf("device uuid: %s not found in sharing configs", devUuid)
}
