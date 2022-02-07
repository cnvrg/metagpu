package deviceplugin

import (
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type DeviceProcessInfo struct {
	Pid          uint32
	GpuMemory    uint64
	Cmdline      []string
	User         string
	ContainerId  string
	PodId        string
	PodNamespace string
}

type MetaDeviceUtilization struct {
	Gpu    uint32
	Memory uint32
}

type DeviceProcess struct {
	Pid       uint32
	GpuMemory uint64
	Process   *process.Process
	ProcFs    *procfs.Proc
}

type MetaDevice struct {
	UUID        string
	Index       int
	Utilization *MetaDeviceUtilization
	Processes   map[uint32]*DeviceProcess
	K8sDevice   *pluginapi.Device
}

type DeviceManager interface {
	CacheDevices()
	ListMetaDevices() []*pluginapi.Device
	DiscoverDeviceProcesses()
	ListDeviceProcesses() map[string][]*DeviceProcessInfo
	ParseRealDeviceId(metaDevicesIds []string) (realDeviceId string)
}
