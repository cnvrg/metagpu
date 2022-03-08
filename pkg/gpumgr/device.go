package gpumgr

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
)

type DeviceMemory struct {
	Total     uint64
	Free      uint64
	Used      uint64
	ShareSize uint64
}

type DeviceUtilization struct {
	Gpu    uint32
	Memory uint32
}

type GpuDevice struct {
	UUID         string
	Index        int
	Shares       int
	ResourceName string
	Utilization  *DeviceUtilization
	Memory       *DeviceMemory
}

func NewGpuDevice(uuid string, index int, utilization nvml.Utilization, memory nvml.Memory) *GpuDevice {
	d := &GpuDevice{
		UUID:        uuid,
		Index:       index,
		Utilization: &DeviceUtilization{Gpu: utilization.Gpu, Memory: utilization.Memory / uint32(MB)},
	}
	// set gpu share configs
	d.setGpuShareConfigs()
	// set gpu memory usage
	d.setGpuMemoryUsage(memory)
	return d
}

func (d *GpuDevice) setGpuShareConfigs() {
	deviceSharingConfigs := sharecfg.NewDeviceSharingConfig()
	if deviceSharing, err := deviceSharingConfigs.GetDeviceSharingConfigs(d.UUID); err != nil {
		log.Fatalf("bad configs, unable to find sharing configs for device: %s", d.UUID)
	} else {
		d.Shares = deviceSharing.MetaGpus
		d.ResourceName = deviceSharing.ResourceName
	}
}

func (d *GpuDevice) setGpuMemoryUsage(memory nvml.Memory) {
	d.Memory = &DeviceMemory{
		Total:     memory.Total / MB,
		Free:      memory.Free / MB,
		Used:      memory.Used / MB,
		ShareSize: memory.Total / uint64(d.Shares) / MB,
	}

}
