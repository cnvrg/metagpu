package sharecfg

import (
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/nvmlutils"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type DeviceSharingConfig struct {
	Uuid           []string
	ResourceName   string
	MetagpusPerGpu int
	AutoReshare    bool
}

type DevicesSharingConfigs struct {
	Configs []*DeviceSharingConfig
}

var shareCfg *DevicesSharingConfigs

func NewDeviceSharingConfig() *DevicesSharingConfigs {
	if shareCfg != nil {
		return shareCfg
	}
	var cfg []*DeviceSharingConfig
	if err := viper.UnmarshalKey("deviceSharing", &cfg); err != nil {
		log.Fatal(err)
	}
	shareCfg = &DevicesSharingConfigs{Configs: cfg}
	shareCfg.ValidateSharingConfiguration()
	shareCfg.AutoReshare()
	return shareCfg
}

func (c *DevicesSharingConfigs) ValidateSharingConfiguration() {
	if len(c.Configs) == 0 {
		log.Fatalf("mission gpu sharing configuration, can't proceed")
	}
	if len(c.Configs) > 1 {
		for _, devCfg := range c.Configs {
			for _, uuid := range devCfg.Uuid {
				if uuid == "*" {
					log.Fatalf("wrong gpu sharing configuration, "+
						"'deviceSharing' with uuid: [ * ] must have sinlge (1) entry, but have: %d", len(c.Configs))
				}
			}
		}
	}
}

func (c *DevicesSharingConfigs) AutoReshare() {
	for _, cfg := range c.Configs {
		if cfg.AutoReshare {
			cfg.GpuAutoResharing()
			continue
		}
		log.Infof("autoReshare disabled for: %s, skipping re-configuration", cfg.ResourceName)
	}
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

func (c *DeviceSharingConfig) GpuAutoResharing() {
	log.Info("autoResharing enabled, re-configuring gpu shares")
	c.MetagpusPerGpu = 100
	// the following code is sharing GPU by memory,
	// currently we are not using it, and I don't think we ever will
	// but, never say never, thus it's here

	//nvmlDevice := c.getFirstDevice()
	//if nvmlDevice != nil {
	//	mem := nvmlutils.GetDeviceMemory(nvmlDevice)
	//	if mem.Total > 0 {
	//		c.MetagpusPerGpu = int((mem.Total / (1024 * 1024)) / 1024)
	//	}
	//}

	// TODO: make sharing configurations persistent
}

// GetShareSize Get share size in MB
func (c *DeviceSharingConfig) GetShareSize() int {
	nvmlDevice := c.getFirstDevice()
	if nvmlDevice != nil {
		mem := nvmlutils.GetDeviceMemory(nvmlDevice)
		if mem.Total > 0 {
			return int((mem.Total / (1024 * 1024)) / uint64(c.MetagpusPerGpu))
		}
	}
	return 0
}

func (c *DeviceSharingConfig) getFirstDevice() *nvml.Device {
	if c.isWildcardSharing() {
		devices := nvmlutils.GetDevices()
		if len(devices) < 0 {
			log.Error("can't execute autoReshare, the devices list is empty")
			return nil
		}
		return devices[0]
	}
	if len(c.Uuid) < 0 {
		log.Error("can't execute autoReshare, uuid config list es empty")
		return nil
	}
	return nvmlutils.GetDeviceByUUID(c.Uuid[0])
}

func (c *DeviceSharingConfig) isWildcardSharing() bool {
	for _, uuid := range c.Uuid {
		if uuid == "*" {
			return true
		}
	}
	return false
}
