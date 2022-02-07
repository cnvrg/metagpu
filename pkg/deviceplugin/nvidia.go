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

type NvidiaDeviceManager struct {
	Devices                  map[string]*MetaDevice
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
}

func (m *NvidiaDeviceManager) CacheDevices() {
	go func() {
		for {
			m.setDevices()
			<-time.After(m.cacheTTL)
		}
	}()
}

func (m *NvidiaDeviceManager) DiscoverDeviceProcesses() {
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
	log.Infof("refreshing nvidia devices cache (total: %d)", count)
	nvmlErrorCheck(ret)
	var discoveredDevices []string
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		uuid, ret := device.GetUUID()
		nvmlErrorCheck(ret)
		discoveredDevices = append(discoveredDevices, uuid)
		if _, ok := m.Devices[uuid]; !ok {
			m.Devices[uuid] = &MetaDevice{
				UUID:      uuid,
				Index:     i,
				Processes: make(map[uint32]*DeviceProcess),
				K8sDevice: &pluginapi.Device{ID: uuid, Health: pluginapi.Healthy},
			}
		}
	}
	// cleanup non-existing devices
	m.cleanUpNonExistingDevices(discoveredDevices)
}

func (m *NvidiaDeviceManager) cleanUpNonExistingDevices(discoveredDevices []string) {
	for deviceUuid, _ := range m.Devices {
		shouldDelete := true
		for _, uuid := range discoveredDevices {
			if deviceUuid == uuid {
				shouldDelete = false
			}
		}
		if shouldDelete {
			delete(m.Devices, deviceUuid)
		}
	}
}

func (m *NvidiaDeviceManager) cleanUpNonExistingDeviceProcesses(deviceUuid string, discoveredProcesses []uint32) {
	var pidToDelete uint32
	shouldDelete := true
	for deviceProcessPid, _ := range m.Devices[deviceUuid].Processes {
		for _, pid := range discoveredProcesses {
			if pid == deviceProcessPid {
				shouldDelete = false
				pidToDelete = pid
			}
		}
	}
	if shouldDelete {
		delete(m.Devices[deviceUuid].Processes, pidToDelete)
	}
}

func (m *NvidiaDeviceManager) discoverGpuProcesses() {
	log.Info("refreshing nvidia devices processes")
	for _, device := range m.Devices {
		nvidiaDevice, ret := nvml.DeviceGetHandleByIndex(device.Index)
		nvmlErrorCheck(ret)
		utilization, ret := nvidiaDevice.GetUtilizationRates()
		nvmlErrorCheck(ret)
		processes, ret := nvidiaDevice.GetComputeRunningProcesses()
		nvmlErrorCheck(ret)
		var discoveredDevicesProcesses []uint32
		for _, nvmlProcessInfo := range processes {
			// check if the discovered pid already exists in device
			if _, ok := device.Processes[nvmlProcessInfo.Pid]; !ok {
				// process does not exist on device, create it
				if dp, err := NewDeviceProcess(nvmlProcessInfo.Pid); err == nil {
					dp.GpuMemory = nvmlProcessInfo.UsedGpuMemory
					device.Processes[nvmlProcessInfo.Pid] = dp
				} else {
					log.Errorf("error creating device process, err: %s", err)
				}
			} else {
				// if process already set, only update the gpu memory
				device.Processes[nvmlProcessInfo.Pid].GpuMemory = nvmlProcessInfo.UsedGpuMemory
			}
			//compose currently running device processes
			discoveredDevicesProcesses = append(discoveredDevicesProcesses, nvmlProcessInfo.Pid)
		}
		// update device utilization
		device.Utilization = &MetaDeviceUtilization{Gpu: utilization.Gpu, Memory: utilization.Memory}
		// cleanup non-existing device processes
		m.cleanUpNonExistingDeviceProcesses(device.UUID, discoveredDevicesProcesses)
	}

	for deviceUuid, deviceProcesses := range m.ListDeviceProcesses() {
		log.Infof("=========== %s ===========", deviceUuid)
		for _, deviceProcess := range deviceProcesses {
			log.Infof("Pid           : %d", deviceProcess.Pid)
			log.Infof("Memory        : %d", deviceProcess.GpuMemory/(1024*1024))
			log.Infof("Command       : %s", deviceProcess.Cmdline)
			log.Infof("ContainerID   : %s", deviceProcess.ContainerId)
			log.Infof("PodName       : %s", deviceProcess.PodId)
			log.Infof("PodNamespace  : %s", deviceProcess.PodNamespace)
		}

	}
}

func (m *NvidiaDeviceManager) ListDeviceProcesses() map[string][]*DeviceProcessInfo {

	deviceProcessInfoMap := make(map[string][]*DeviceProcessInfo)
	for uuid, device := range m.Devices {
		var deviceProcesses []*DeviceProcessInfo
		for pid, deviceProcess := range device.Processes {
			var e error
			// set pid and gpu memory
			pInfo := &DeviceProcessInfo{Pid: pid, GpuMemory: deviceProcess.GpuMemory}
			// discover cmdline
			pInfo.Cmdline, e = deviceProcess.Process.CmdlineSlice()
			checkProcessDiscoveryError(e)
			// discover username
			pInfo.User, e = deviceProcess.Process.Username()
			checkProcessDiscoveryError(e)
			// getting cgroups for discovering containerId
			var cgroups []procfs.Cgroup
			cgroups, e = deviceProcess.ProcFs.Cgroups()
			if len(cgroups) == 0 {
				log.Errorf("cgroups list for %d is empty", pid)
			} else {
				// extract pod name and pod namespace from container labels
				pInfo.ContainerId = filepath.Base(cgroups[0].Path)
				pInfo.PodId, pInfo.PodNamespace = inspectContainer(pInfo.ContainerId)
			}
			deviceProcesses = append(deviceProcesses, pInfo)
		}
		deviceProcessInfoMap[uuid] = deviceProcesses
	}
	return deviceProcessInfoMap
}

func NewNvidiaDeviceManager() *NvidiaDeviceManager {
	ret := nvml.Init()
	nvmlErrorCheck(ret)
	ndm := &NvidiaDeviceManager{
		Devices:                  make(map[string]*MetaDevice),
		cacheTTL:                 time.Second * time.Duration(viper.GetInt("deviceCacheTTL")),
		processesDiscoveryPeriod: time.Second * time.Duration(viper.GetInt("processesDiscoveryPeriod")),
	}
	ndm.CacheDevices()
	ndm.DiscoverDeviceProcesses()
	return ndm
}

func NewDeviceProcess(pid uint32) (*DeviceProcess, error) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}
	procFs, err := procfs.NewProc(int(pid))
	if err != nil {
		return nil, err
	}
	return &DeviceProcess{
		Pid:     pid,
		Process: proc,
		ProcFs:  &procFs,
	}, nil
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
	defer cli.Close()
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
