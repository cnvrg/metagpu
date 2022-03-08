package sharecfg

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
	if err := viper.UnmarshalKey("deviceSharing", &cfg); err != nil {
		log.Fatal(err) // TODO: add context to logs!!!!
	}
	if len(cfg) > 1 {
		for _, devCfg := range cfg {
			for _, uuid := range devCfg.Uuid {
				if uuid == "*" {
					log.Fatalf("wrong gpu sharing configuration, "+
						"'deviceSharing' with uuid: [ * ] must have sinlge (1) entry, but have: %d", len(cfg))
				}
			}
		}
	}
	return &DevicesSharingConfigs{Configs: cfg}
}

func (c *DevicesSharingConfigs) GetDeviceSharingConfigs(devUuid string) (*DeviceSharingConfig, error) {
	for _, devCfg := range c.Configs {
		for _, uuid := range devCfg.Uuid {
			if uuid == devUuid || uuid == "*" {
				return devCfg, nil
			}
		}
	}
	return nil, fmt.Errorf("device uuid: %s not found in sharing configs", devUuid)
}
