package plugin

import (
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"time"
)

type DeviceManager interface {
	GetPluginDevices(gpuShares int) []*pluginapi.Device
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize, totalShares int, availableDevIds []string) ([]string, error)
}

type DeviceUuid string

type MetaGpuDevicePlugin struct {
	DeviceManager
	server               *grpc.Server
	socket               string
	resourceName         string
	deviceUuids          []string
	totalShares          int
	stop                 chan interface{}
	MetaGpuRecalculation chan bool
}

type NvidiaDeviceManager struct {
	Devices                  []*MetaDevice
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
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
