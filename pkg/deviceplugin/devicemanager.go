package deviceplugin

import pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

type DeviceProcess struct {
	Pid          uint32
	Memory       uint64
	Cmdline      string
	User         string
	ContainerId  string
	podId        string
	podNamespace string
}

type DeviceManager interface {
	CacheDevices()
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
