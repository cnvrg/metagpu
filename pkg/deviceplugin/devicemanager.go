package deviceplugin

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

//
//type MetaDeviceInfo struct {
//	Node     string
//	Metadata map[string]string
//	Devices  []*MetaDevice
//}

type DeviceManager interface {
	//CacheDevices()
	//DiscoverDeviceProcesses()
	//GetMetaDeviceInfo() *MetaDeviceInfo
	//GetMetaDevices() map[string]*MetaDevice
	GetPluginDevices() []*pluginapi.Device
	//GetGpuShareMemSize(uuid string) (shareSize uint64)
	//GetProcesses(podId string) []*DeviceProcess
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize, totalShares int, availableDevIds []string) ([]string, error)
	//KillGpuProcess(pid uint32) error
}
