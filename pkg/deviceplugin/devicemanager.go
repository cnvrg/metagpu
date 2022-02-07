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
	ListDeviceProcesses() map[string][]*DeviceProcess
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
