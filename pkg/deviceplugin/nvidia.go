package deviceplugin

import (
	"context"
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	dockerclient "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

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
	// enforce device discovery
	// to make sure all the devices will be set
	// before kubelet api server will be started
	m.setDevices()
	go func() {
		for {
			<-time.After(m.cacheTTL)
			m.setDevices()
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
	log.Infof("generating meta gpu devices (total: %d)", len(m.Devices)*viper.GetInt("metaGpus"))
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
	discoveredDevices := make(map[string]*MetaDevice)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		uuid, ret := device.GetUUID()
		nvmlErrorCheck(ret)
		discoveredDevices[uuid] = &MetaDevice{
			UUID:      uuid,
			Index:     i,
			K8sDevice: &pluginapi.Device{ID: uuid, Health: pluginapi.Healthy},
		}
	}
	m.Devices = discoveredDevices
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
		var discoveredDevicesProcesses []*DeviceProcess
		for _, nvmlProcessInfo := range processes {
			discoveredDevicesProcesses = append(discoveredDevicesProcesses,
				NewDeviceProcess(nvmlProcessInfo.Pid, nvmlProcessInfo.UsedGpuMemory))
		}
		// override device utilization
		device.Processes = discoveredDevicesProcesses
		device.Utilization = &DeviceUtilization{Gpu: utilization.Gpu, Memory: utilization.Memory}
	}

	for deviceUuid, deviceProcesses := range m.ListDeviceProcesses() {
		log.Infof("=========== %s ===========", deviceUuid)
		for _, deviceProcess := range deviceProcesses {
			log.Infof("Pid           : %d", deviceProcess.Pid)
			log.Infof("Memory        : %d", deviceProcess.GpuMemory/(1024*1024))
			log.Infof("Command       : %s", deviceProcess.GetShortCmdLine())
			log.Infof("ContainerID   : %s", deviceProcess.ContainerId)
			log.Infof("PodName       : %s", deviceProcess.PodId)
			log.Infof("PodNamespace  : %s", deviceProcess.PodNamespace)
			log.Info("--------")
		}
	}
}

func (m *NvidiaDeviceManager) ListDeviceProcesses() map[string][]*DeviceProcess {

	deviceProcessInfoMap := make(map[string][]*DeviceProcess)
	for uuid, device := range m.Devices {
		deviceProcessInfoMap[uuid] = device.Processes
	}
	return deviceProcessInfoMap
}

func nvmlErrorCheck(ret nvml.Return) {
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
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

func NewNvidiaDeviceManager() *NvidiaDeviceManager {
	ret := nvml.Init()
	nvmlErrorCheck(ret)
	ndm := &NvidiaDeviceManager{
		Devices:                  make(map[string]*MetaDevice),
		cacheTTL:                 time.Second * time.Duration(viper.GetInt("deviceCacheTTL")),
		processesDiscoveryPeriod: time.Second * time.Duration(viper.GetInt("processesDiscoveryPeriod")),
	}

	// start cache devices loop
	ndm.CacheDevices()
	// start process discovery loop
	ndm.DiscoverDeviceProcesses()
	return ndm
}
