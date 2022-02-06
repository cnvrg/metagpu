package deviceplugin

import pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

type DeviceProcess struct {
	pid          uint32
	memory       uint64
	cmdline      string
	user         string
	containerId  string
	podId        string
	podNamespace string
}

type DeviceManager interface {
	CacheDevices()
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
