package plugin

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"time"
)

type DeviceManager interface {
	GetPluginDevices() []*pluginapi.Device
	GetDeviceSharingConfig() *sharecfg.DeviceSharingConfig
	GetUnixSocket() string
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error)
}

type DeviceUuid string

type MetaGpuDevicePlugin struct {
	DeviceManager
	server               *grpc.Server
	socket               string
	stop                 chan interface{}
	MetaGpuRecalculation chan bool
}

type NvidiaDeviceManager struct {
	Devices                  []*MetaDevice
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
	shareCfg                 *sharecfg.DeviceSharingConfig
}

type MetaDevice struct {
	UUID  string
	Index int
}

type DeviceLoad struct {
	Metagpus []string
}

type DeviceAllocation struct {
	LoadMap             []*DeviceLoad
	AvailableDevIds     []string
	AllocationSize      int
	TotalShares         int
	MetagpusAllocations []string
}
