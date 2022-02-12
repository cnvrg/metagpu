package deviceplugin

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type DeviceUtilization struct {
	Gpu    uint32
	Memory uint32
}

type MetaDevice struct {
	UUID        string
	Index       int
	Utilization *DeviceUtilization
	Processes   []*DeviceProcess
	K8sDevice   *pluginapi.Device
}

type DeviceManager interface {
	CacheDevices()
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ListDeviceProcesses(podId string) map[DeviceUuid][]*DeviceProcess
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId []string)
	MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error)
	KillGpuProcess(pid uint32) error
}
