package deviceplugin

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
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

type MetaDevice struct {
	UUID        string
	Index       int
	Shares      int
	Utilization *DeviceUtilization
	Memory      *DeviceMemory
}

type DeviceManager interface {
	CacheDevices()
	DiscoverDeviceProcesses()
	ListDevices() map[string]*MetaDevice
	ListMetaDevices() []*pluginapi.Device
	GetGpuShareMemSize(uuid string) (shareSize uint64)
	ListProcesses(podId string) []*DeviceProcess
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error)
	KillGpuProcess(pid uint32) error
}
