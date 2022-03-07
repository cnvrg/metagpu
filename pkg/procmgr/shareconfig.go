package procmgr

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type DeviceSharing struct {
	Uuid         string
	ResourceName string
	MetaGpus     int
}

type DevicesSharesConfigs struct {
	DeviceShare map[string]*DeviceSharing
}

func NewDeviceSharingConfig() *DevicesSharesConfigs {
	var cfg []*DeviceSharing
	sharingConfigs := &DevicesSharesConfigs{DeviceShare: make(map[string]*DeviceSharing)}
	if err := viper.UnmarshalKey("deviceSharing", cfg); err != nil {
		log.Fatal(err)
	}
	for _, c := range cfg {
		sharingConfigs.DeviceShare[c.Uuid] = c
	}
	return sharingConfigs
}

func (c *DevicesSharesConfigs) getDeviceSharingConfigs(devUuid string) (*DeviceSharing, error) {
	if _, ok := c.DeviceShare[devUuid]; !ok {
		return nil, fmt.Errorf("device uuid: %s not found in sharing configs", devUuid)
	}
	return c.DeviceShare[devUuid], nil
}
