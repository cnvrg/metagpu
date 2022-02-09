package deviceplugin

import (
	"errors"
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
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

func (m *NvidiaDeviceManager) ParseRealDeviceId(metaDevicesIds []string) (realDevicesIds []string) {

	// each meta gpu will start from 'cnvrg-meta-[index-number]-[sequence-number]'
	r, _ := regexp.Compile("cnvrg-meta-\\d+-\\d+-")
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
	return realDevicesIdsList
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
				ID:     fmt.Sprintf("cnvrg-meta-%d-%d-%s", d.Index, j, d.K8sDevice.ID),
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

	_ = getMetagpuAnonymouseWorkloads()
	//for _, device := range m.Devices {
	//	for _, deviceProcess := range device.Processes {
	//		podFound := false
	//		for _, metaGpuPod := range gpuEnabledPods {
	//			for _, container := range pod.Spec.Containers {
	//
	//			}
	//		}
	//	}
	//}

	for deviceUuid, deviceProcesses := range m.ListDeviceProcesses() {
		log.Infof("=========== %s ===========", deviceUuid)
		for _, deviceProcess := range deviceProcesses {
			log.Infof("Pid             : %d", deviceProcess.Pid)
			log.Infof("Memory          : %d", deviceProcess.GpuMemory/(1024*1024))
			log.Infof("Command         : %s", deviceProcess.GetShortCmdLine())
			log.Infof("ContainerID     : %s", deviceProcess.ContainerId)
			log.Infof("PodName         : %s", deviceProcess.PodId)
			log.Infof("PodNamespace    : %s", deviceProcess.PodNamespace)
			log.Infof("MetagpuRequest  : %d", deviceProcess.PodMetagpuRequest)
			log.Info("--------")
		}
	}
	log.Info("************************************")
}

func (m *NvidiaDeviceManager) ListDeviceProcesses() map[string][]*DeviceProcess {

	deviceProcessInfoMap := make(map[string][]*DeviceProcess)
	for uuid, device := range m.Devices {
		deviceProcessInfoMap[uuid] = device.Processes
	}

	return deviceProcessInfoMap
}

func (m *NvidiaDeviceManager) MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error) {

	// get total shares per gpu
	totalSharesPerGPU := viper.GetInt("metaGpus")
	// detect device load
	deviceLoad := NewDeviceLoadMap(m.ParseRealDeviceId(availableDevIds), availableDevIds)
	// calculate how entire and how shares are required
	entireGpusRequest := allocationSize / totalSharesPerGPU
	gpuFractionsRequest := allocationSize % totalSharesPerGPU
	log.Infof("metagpu allocation request: %d.%d", entireGpusRequest, gpuFractionsRequest)
	// detect entirely allocatable gpus
	allocatableGPUs, err := findEntirelyAllocatableGPUs(entireGpusRequest, deviceLoad)
	if err != nil {
		return nil, err
	}
	// detect fractional allocatable gpus
	allocatableGPUs, err = findFractionalAllocatableGPUs(gpuFractionsRequest, deviceLoad, allocatableGPUs)
	if err != nil {
		return nil, err
	}
	// compose the device comma seperated string and return to K8s Allocation
	return composeDevUuidsString(append(entirelyAllocatableGPUs, partialAllocatableGPUs...)), nil
}

func nvmlErrorCheck(ret nvml.Return) {
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}

func findEntirelyAllocatableGPUs(quantity int, deviceLoad *DeviceAllocationMap) (allocatedDevices map[string]int, e error) {
	allocatedDevices = make(map[string]int)
	totalSharesPerGPU := viper.GetInt("metaGpus")
	if quantity == 0 {
		return
	}
	for devUuid, load := range deviceLoad.LoadMap {
		if load.FreeShares == totalSharesPerGPU { // meaning gpu is entirely free
			allocatedDevices[devUuid] = totalSharesPerGPU // allocate all gpu shares to single gpu
		}
		// once we got enough entirely free gpus, break the loop
		if len(allocatedDevices) == quantity {
			break
		}
	}
	if len(allocatedDevices) < quantity {
		return nil, errors.New("can't allocate entirely requested gpus quantity")
	}

	for devUuid, _ := range allocatedDevices {
		deviceLoad.MetagpusAllocations = append(deviceLoad.MetagpusAllocations, deviceLoad.LoadMap[devUuid].Metagpus...)
	}
	return
}

func findFractionalAllocatableGPUs(quantity int, deviceLoad *DeviceAllocationMap, allocatableGPUs map[string]int) (allocatedDevices map[string]int, e error) {

	// find free gpu fraction and allocate them
	for devUuid, load := range deviceLoad.LoadMap {
		if _, ok := allocatableGPUs[devUuid]; !ok && load.FreeShares >= quantity {
			allocatableGPUs[devUuid] = quantity
			break
		}
	}
	if len(allocatedDevices) == 0 {
		return nil, errors.New("can't allocate requested gpu shares")
	}
	return
}

func buildDevicesLoadMap(realDeviceIds []string, availableDevIds []string) map[string]int {
	var deviceLoad = make(map[string]int)
	var deviceToMetagpus = make(map[string][]string)
	for _, deviceId := range realDeviceIds {
		for _, availableDevId := range availableDevIds {
			if strings.Contains(availableDevId, deviceId) {
				deviceToMetagpus[deviceId] = append(deviceToMetagpus[deviceId], availableDevId)
			}
		}
	}
	for devId, metagpus := range deviceToMetagpus {
		deviceLoad[devId] = len(metagpus)
	}
	return deviceLoad
}

//func gpuToMetagpu(allocatedDevices map[string]int) []string {
//	var devUuidSet []string
//	devUuids := map[string]bool{}
//
//	// eliminates duplicates
//	for _, devUuid := range uuids {
//		devUuids[devUuid] = true
//	}
//	// compose list without duplicates
//	for devUuid, _ := range devUuids {
//		devUuidSet = append(devUuidSet, devUuid)
//	}
//	// convert to string and return
//	return strings.Join(devUuidSet, ",")
//}

//func containsString(slice []string, s string) bool {
//	for _, item := range slice {
//		if item == s {
//			return true
//		}
//	}
//	return false
//}
//func containsStringInMapKey(slice map[string]int, s string) bool {
//	for _, item := range slice {
//		if item == s {
//			return true
//		}
//	}
//	return false
//}

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
