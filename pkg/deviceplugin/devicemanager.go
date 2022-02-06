package deviceplugin

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type MetaDeviceUtilization struct {
	Gpu    uint32
	Memory uint32
}

type DeviceProcess struct {
	Pid          uint32
	Memory       uint64
	Cmdline      string
	User         string
	ContainerId  string
	podId        string
	podNamespace string
}

type MetaDevice struct {
	UUID        string
	Index       int
	Utilization *MetaDeviceUtilization
	Processes   []*DeviceProcess
	K8sDevice   *pluginapi.Device
}

type DeviceManager interface {
	CacheDevices()
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ListCachedDeviceProcesses() []*MetaDevice
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
