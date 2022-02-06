package deviceplugin

import (
	"context"
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	dockerclient "github.com/docker/docker/client"
	"github.com/prometheus/procfs"
	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//type MetaDevice struct {
//	K8sDevice   *pluginapi.Device
//	NDevice     *nvml.Device
//	Utilization *nvml.Utilization
//	Processes   []*DeviceProcess
//}

type NvidiaDeviceManager struct {
	Devices                  []*MetaDevice
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
}

func (m *NvidiaDeviceManager) CacheDevices() {
	m.setDevices()
	go func() {
		for {
			<-time.After(m.cacheTTL)
			m.setDevices()
		}
	}()
}

func (m *NvidiaDeviceManager) DiscoverDeviceProcesses() {
	m.CacheDevices()
	go func() {
		for {
			m.discoverGpuProcesses()
			<-time.After(m.processesDiscoveryPeriod)
		}
	}()
}

func (m *NvidiaDeviceManager) ParseRealDeviceId(metaDevicesIds []string) (realDevicesIds string) {

	// each meta gpu will starts from 'cnvrg-meta-[number]-'
	r, _ := regexp.Compile("cnvrg-meta-\\d+-")
	// string map will eliminate doubles in real Devices ids
	realDevicesIdsMap := make(map[string]bool)
	for _, metaDeviceId := range metaDevicesIds {
		deviceId := r.ReplaceAllString(metaDeviceId, "")
		if !m.DeviceExists(deviceId) {
			log.Errorf("device %s doesn not exists, but was claimed", metaDeviceId)
			continue
		}
		realDevicesIdsMap[deviceId] = true
	}

	var realDevicesIdsList []string
	for dId, _ := range realDevicesIdsMap {
		realDevicesIdsList = append(realDevicesIdsList, dId)
	}
	// TODO: verify list is not empty!
	realDevicesIds = strings.Join(realDevicesIdsList, ",")
	if len(realDevicesIds) == 0 {
		realDevicesIds = "none"
	}
	return realDevicesIds
}

func (m *NvidiaDeviceManager) DeviceExists(deviceId string) bool {
	for _, d := range m.Devices {
		if d.K8sDevice.ID == deviceId {
			return true
		}
	}
	return false
}

func (m *NvidiaDeviceManager) ListMetaDevices() []*pluginapi.Device {
	var metaGpus []*pluginapi.Device
	log.Infof("generating meta gpu Devices (total: %d)", len(m.Devices)*viper.GetInt("metaGpus"))
	for _, d := range m.Devices {
		for j := 0; j < viper.GetInt("metaGpus"); j++ {
			metaGpus = append(metaGpus, &pluginapi.Device{
				ID:     fmt.Sprintf("cnvrg-meta-%d-%s", j, d.K8sDevice.ID),
				Health: pluginapi.Healthy,
			})
		}
	}

	return metaGpus

}

func (m *NvidiaDeviceManager) setDevices() {

	count, ret := nvml.DeviceGetCount()
	log.Infof("refreshing nvidia Devices cache (total: %d)", count)
	nvmlErrorCheck(ret)
	var dl []*MetaDevice
	for i := 0; i < count; i++ {
		var md MetaDevice
		device, ret := nvml.DeviceGetHandleByIndex(i)
		uuid, ret := device.GetUUID()
		nvmlErrorCheck(ret)
		md.K8sDevice = &pluginapi.Device{ID: uuid, Health: pluginapi.Healthy}
		md.UUID = uuid
		md.Index = i
		nvmlErrorCheck(ret)
		dl = append(dl, &md)
	}
	m.Devices = dl
}

func (m *NvidiaDeviceManager) ListCachedDeviceProcesses() []*MetaDevice {
	var metaDeviceList []*MetaDevice
	for _, d := range m.Devices {
		metaDeviceList = append(metaDeviceList, &MetaDevice{
			UUID:        d.UUID,
			Index:       d.Index,
			Utilization: d.Utilization,
			Processes:   d.Processes,
			K8sDevice:   d.K8sDevice,
		})
	}
	return metaDeviceList
}

func (m *NvidiaDeviceManager) discoverGpuProcesses() {
	for _, device := range m.Devices {
		nvidiaDevice, ret := nvml.DeviceGetHandleByIndex(device.Index)
		nvmlErrorCheck(ret)
		utilization, ret := nvidiaDevice.GetUtilizationRates()
		nvmlErrorCheck(ret)
		processes, ret := nvidiaDevice.GetComputeRunningProcesses()
		nvmlErrorCheck(ret)
		var processList []*DeviceProcess
		for _, nvmlProcessInfo := range processes {
			gpuProcess := DeviceProcess{Pid: nvmlProcessInfo.Pid, Memory: nvmlProcessInfo.UsedGpuMemory}
			gpuProcess.enrichProcessInfo()
			processList = append(processList, &gpuProcess)
		}
		device.Processes = processList
		device.Utilization = &MetaDeviceUtilization{
			Gpu:    utilization.Gpu,
			Memory: utilization.Memory,
		}
	}
	for _, device := range m.Devices {
		log.Infof("=========== %s ===========", device.K8sDevice.ID)
		for _, p := range device.Processes {
			cmd := ""
			if p.Cmdline != "" {
				cmd = strings.Split(p.Cmdline, " ")[0]
			}

			log.Infof("Pid           : %d", p.Pid)
			log.Infof("Memory        : %d", p.Memory/(1024*1024))
			log.Infof("Command       : %s", cmd)
			log.Infof("ContainerID   : %s", p.ContainerId)
			log.Infof("PodName       : %s", p.podId)
			log.Infof("PodNamespace  : %s", p.podNamespace)
		}
		log.Info("=========================")
	}
}

func (p *DeviceProcess) enrichProcessInfo() {

	if pr, err := process.NewProcess(int32(p.Pid)); err == nil {
		var e error
		var cmdline, user string
		cmdline, e = pr.Cmdline()
		checkProcessDiscoveryError(e)
		user, e = pr.Username()
		checkProcessDiscoveryError(e)
		p.Cmdline = cmdline
		p.User = user

	} else {
		log.Error(err)
	}

	if proc, err := procfs.NewProc(int(p.Pid)); err == nil {
		var e error
		var cgroups []procfs.Cgroup
		cgroups, e = proc.Cgroups()
		if e != nil {
			log.Error(e)
		}
		if len(cgroups) == 0 {
			log.Errorf("cgroups list for %d is empty", p.Pid)
		}
		p.ContainerId = filepath.Base(cgroups[0].Path)
		p.podId, p.podNamespace = inspectContainer(p.ContainerId)

	}
}

func NewNvidiaDeviceManager() *NvidiaDeviceManager {
	ret := nvml.Init()
	nvmlErrorCheck(ret)
	ndm := &NvidiaDeviceManager{
		cacheTTL:                 time.Second * time.Duration(viper.GetInt("deviceCacheTTL")),
		processesDiscoveryPeriod: time.Second * time.Duration(viper.GetInt("processesDiscoveryPeriod")),
	}
	ndm.CacheDevices()
	ndm.DiscoverDeviceProcesses()
	return ndm
}

func nvmlErrorCheck(ret nvml.Return) {
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}

func checkProcessDiscoveryError(e error) {
	if e != nil {
		log.Error(e)
	}
}

func inspectContainer(containerId string) (podName, podNamespace string) {

	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		log.Error(err)
		return
	}
	cd, err := cli.ContainerInspect(context.Background(), containerId)
	if err != nil {
		log.Error(err)
		return
	}
	if pd, ok := cd.Config.Labels["io.kubernetes.pod.name"]; ok {
		podName = pd
	}

	if pn, ok := cd.Config.Labels["io.kubernetes.pod.namespace"]; ok {
		podNamespace = pn
	}

	return
}
