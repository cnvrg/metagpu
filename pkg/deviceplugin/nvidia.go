package deviceplugin

import (
	"fmt"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"regexp"
	"time"
)

type NvidiaDeviceManager struct {
	Devices                  []*MetaDevice
	Processes                []*DeviceProcess
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
}

var MB uint64 = 1024 * 1024

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
			m.discoverGpuProcessesAndDevicesLoad()
			<-time.After(m.processesDiscoveryPeriod)
		}
	}()
}

func (m *NvidiaDeviceManager) GetGpuShareMemSize(uuid string) (shareSize uint64) {
	for _, d := range m.Devices {
		if d.UUID == uuid {
			return d.Memory.ShareSize
		}
	}
	return
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
		if d.UUID == deviceId {
			return true
		}
	}
	return false
}

func (m *NvidiaDeviceManager) ListMetaDevices() []*pluginapi.Device {
	var metaGpus []*pluginapi.Device
	metaGpusQuantity := viper.GetInt("metaGpus")
	log.Infof("generating meta gpu devices (total: %d)", len(m.Devices)*metaGpusQuantity)
	for _, d := range m.Devices {
		for j := 0; j < metaGpusQuantity; j++ {
			metaGpus = append(metaGpus, &pluginapi.Device{
				ID:     fmt.Sprintf("cnvrg-meta-%d-%d-%s", d.Index, j, d.UUID),
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
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		uuid, ret := device.GetUUID()
		nvmlErrorCheck(ret)
		m.Devices = append(m.Devices, &MetaDevice{UUID: uuid, Index: i})
	}
}

func (m *NvidiaDeviceManager) discoverGpuProcessesAndDevicesLoad() {
	var discoveredDevicesProcesses []*DeviceProcess
	for _, device := range m.Devices {
		nvidiaDevice, ret := nvml.DeviceGetHandleByIndex(device.Index)
		nvmlErrorCheck(ret)
		deviceMemory, ret := nvidiaDevice.GetMemoryInfo()
		nvmlErrorCheck(ret)
		utilization, ret := nvidiaDevice.GetUtilizationRates()
		nvmlErrorCheck(ret)
		processes, ret := nvidiaDevice.GetComputeRunningProcesses()
		nvmlErrorCheck(ret)
		for _, nvmlProcessInfo := range processes {
			discoveredDevicesProcesses = append(discoveredDevicesProcesses,
				NewDeviceProcess(nvmlProcessInfo.Pid, nvmlProcessInfo.UsedGpuMemory/MB, device.UUID))
		}
		device.Utilization = &DeviceUtilization{Gpu: utilization.Gpu, Memory: utilization.Memory / uint32(MB)}
		device.Memory = &DeviceMemory{
			Total:     deviceMemory.Total / MB,
			Free:      deviceMemory.Free / MB,
			Used:      deviceMemory.Used / MB,
			ShareSize: deviceMemory.Total / viper.GetUint64("metaGpus") / MB,
		}
		device.Shares = viper.GetInt("metaGpus")
	}
	m.Processes = discoveredDevicesProcesses
}

func (m *NvidiaDeviceManager) ListDevices() map[string]*MetaDevice {
	var deviceMap = make(map[string]*MetaDevice)
	for _, d := range m.Devices {
		deviceMap[d.UUID] = d
	}
	return deviceMap
}

func (m *NvidiaDeviceManager) ListProcesses(podId string) []*DeviceProcess {

	if podId != "" {
		var podProcesses []*DeviceProcess
		for _, deviceProcess := range m.Processes {
			if deviceProcess.PodId == podId {
				podProcesses = append(podProcesses, deviceProcess)
			}
		}
		return podProcesses
	}
	return m.Processes
}

func (m *NvidiaDeviceManager) KillGpuProcess(pid uint32) error {
	p := NewDeviceProcess(pid, 0, "")
	return p.Kill()
}

func (m *NvidiaDeviceManager) MetagpuAllocation(allocationSize int, availableDevIds []string) ([]string, error) {
	return NewDeviceAllocation(allocationSize, availableDevIds).MetagpusAllocations, nil
}

func nvmlErrorCheck(ret nvml.Return) {
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}

func NewNvidiaDeviceManager() *NvidiaDeviceManager {
	ret := nvml.Init()
	nvmlErrorCheck(ret)
	ndm := &NvidiaDeviceManager{
		cacheTTL:                 time.Second * time.Duration(viper.GetInt("deviceCacheTTL")),
		processesDiscoveryPeriod: time.Second * time.Duration(viper.GetInt("processesDiscoveryPeriod")),
	}

	// start cache devices loop
	ndm.CacheDevices()
	// start process discovery loop
	ndm.DiscoverDeviceProcesses()
	return ndm
}
