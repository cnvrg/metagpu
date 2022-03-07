package plugin

import (
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/nvmlutils"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"regexp"
	"time"
)

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

func (m *NvidiaDeviceManager) GetPluginDevices(metaGpusQuantity int) []*pluginapi.Device {
	var metaGpus []*pluginapi.Device
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
	nvmlutils.ErrorCheck(ret)
	var devices []*MetaDevice
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		uuid, ret := device.GetUUID()
		// enable accounting mode
		nvmlutils.ErrorCheck(ret)
		ret = device.SetAccountingMode(nvml.FEATURE_ENABLED)
		nvmlutils.ErrorCheck(ret)
		// verify accounting mode is on,
		// seems like for MIG-enabled devices we can't enable accounting mode?
		// https://github.com/NVIDIA/nvidia-settings/blob/main/src/nvml.h#L5717
		state, ret := device.GetAccountingMode()
		nvmlutils.ErrorCheck(ret)
		log.Infof("accounting mode for device: %s is: %d", uuid, state)
		devices = append(devices, &MetaDevice{UUID: uuid, Index: i})
	}
	m.Devices = devices
}

func (m *NvidiaDeviceManager) MetagpuAllocation(allocationSize, totalShares int, availableDevIds []string) ([]string, error) {
	return NewDeviceAllocation(allocationSize, totalShares, availableDevIds).MetagpusAllocations, nil
}

func NewNvidiaDeviceManager() *NvidiaDeviceManager {
	ret := nvml.Init()
	nvmlutils.ErrorCheck(ret)
	ndm := &NvidiaDeviceManager{
		cacheTTL:                 time.Second * time.Duration(viper.GetInt("deviceCacheTTL")),
		processesDiscoveryPeriod: time.Second * time.Duration(viper.GetInt("processesDiscoveryPeriod")),
	}
	// start cache devices loop
	ndm.CacheDevices()
	return ndm
}
