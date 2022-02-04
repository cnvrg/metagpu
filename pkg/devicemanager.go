package pkg

import pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

type DeviceProcess struct {
	pid         uint32
	memory      uint64
	cmdline     string
	user        string
	containerId string
	podId       string
}

type DeviceManager interface {
	CacheDevices()
	ListDevices() []*pluginapi.Device
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
