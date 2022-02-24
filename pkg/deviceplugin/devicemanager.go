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

type MetaDeviceInfo struct {
	Node     string
	Metadata map[string]string
	Devices  []*MetaDevice
}

type DeviceManager interface {
	CacheDevices()
	DiscoverDeviceProcesses()
	GetMetaDeviceInfo() *MetaDeviceInfo
	GetMetaDevices() map[string]*MetaDevice
	GetPluginDevices() []*pluginapi.Device
	GetGpuShareMemSize(uuid string) (shareSize uint64)
	GetProcesses(podId string) []*DeviceProcess
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error)
	KillGpuProcess(pid uint32) error
}
