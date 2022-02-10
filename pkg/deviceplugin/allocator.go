package deviceplugin

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"regexp"
	"sort"
	"strings"
)

type DeviceUuid string

type DeviceLoad struct {
	Metagpus []string
}

type DeviceAllocation struct {
	LoadMap             map[DeviceUuid]*DeviceLoad
	AvailableDevIds     []string
	AllocationSize      int
	TotalSharesPerGPU   int
	MetagpusAllocations []string
}

func NewDeviceAllocation(allocationSize int, availableDevIds []string) *DeviceAllocation {
	sort.Strings(availableDevIds)
	devAlloc := &DeviceAllocation{
		AvailableDevIds:   availableDevIds,
		AllocationSize:    allocationSize,
		TotalSharesPerGPU: viper.GetInt("metaGpus"),
	}
	devAlloc.InitLoadMap()
	devAlloc.SetAllocations()
	return devAlloc
}

func (a *DeviceAllocation) InitLoadMap() {
	a.LoadMap = make(map[DeviceUuid]*DeviceLoad)
	// build a map of real device id to meta device id
	for _, deviceId := range a.MetaDeviceIdsToRealDeviceIds() {
		for _, availableDevId := range a.AvailableDevIds {
			if !strings.Contains(availableDevId, deviceId) {
				continue
			}
			if a.LoadMap[DeviceUuid(deviceId)] == nil {
				a.LoadMap[DeviceUuid(deviceId)] = &DeviceLoad{}
			}
			a.LoadMap[DeviceUuid(deviceId)].Metagpus = append(a.LoadMap[DeviceUuid(deviceId)].Metagpus, availableDevId)
		}
	}
}

func (a *DeviceAllocation) MetaDeviceIdsToRealDeviceIds() (realDeviceIds []string) {
	// each meta gpu will start from 'cnvrg-meta-[index-number]-[sequence-number]'
	r, _ := regexp.Compile("cnvrg-meta-\\d+-\\d+-")
	// map[string] will eliminate doubles in real Devices ids
	realDevicesIdsMap := make(map[string]bool)
	for _, metaDeviceId := range a.AvailableDevIds {
		realDevicesIdsMap[r.ReplaceAllString(metaDeviceId, "")] = true
	}

	for deviceId, _ := range realDevicesIdsMap {
		realDeviceIds = append(realDeviceIds, deviceId)
	}

	return
}

func (a *DeviceAllocation) SetAllocations() {
	entireGpusRequest := a.AllocationSize / a.TotalSharesPerGPU
	gpuFractionsRequest := a.AllocationSize % a.TotalSharesPerGPU
	log.Infof("metagpu allocation request: %d.%d", entireGpusRequest, gpuFractionsRequest)

	// first try to allocate entire gpus if requested
	for i := 0; i < entireGpusRequest; i++ {
		for _, devLoad := range a.LoadMap {
			if devLoad.getFreeShares() == a.TotalSharesPerGPU {
				a.MetagpusAllocations = append(a.MetagpusAllocations, devLoad.Metagpus...)
				devLoad.Metagpus = nil
			}
		}
	}

	if gpuFractionsRequest > 0 {
		for _, devLoad := range a.LoadMap {
			if devLoad.getFreeShares() >= gpuFractionsRequest {
				var devicesToAdd []string
				for i, device := range devLoad.Metagpus {
					if i == gpuFractionsRequest {
						break
					}
					devicesToAdd = append(devicesToAdd, device)
				}
				a.MetagpusAllocations = append(a.MetagpusAllocations, devicesToAdd...)
				devLoad.removeDevices(devicesToAdd)
				break
			}
		}
		// if still missing allocations,
		// meaning wasn't able to allocate required fractions from the same GPU
		// will try to allocate a fractions from different GPUs
		if len(a.MetagpusAllocations) != a.AllocationSize {
			allocationsLeft := a.AllocationSize
			for _, devLoad := range a.LoadMap {
				for _, device := range devLoad.Metagpus {
					a.MetagpusAllocations = append(a.MetagpusAllocations, device)
					allocationsLeft--
					if allocationsLeft == 0 {
						break
					}
				}
				if allocationsLeft == 0 {
					break
				}
			}
		}
	}
	if len(a.MetagpusAllocations) != a.AllocationSize {
		log.Errorf("error during allocation, the allocationSize: %d doesn't match total allocated devices: %d", a.AllocationSize, len(a.MetagpusAllocations))
	}
}

func (l *DeviceLoad) getFreeShares() int {
	return len(l.Metagpus)
}

func (l *DeviceLoad) removeDevices(devIds []string) {
	for _, devId := range devIds {
		for i, v := range l.Metagpus {
			if v == devId {
				l.Metagpus = append(l.Metagpus[:i], l.Metagpus[i+1:]...)
			}
		}
	}
}

//func (m *DeviceAllocation) AllocateMetagpus(allocatableGPUs map[string]int) []string {
//	for devUuid, size := range allocatableGPUs {
//		for i = 0; i < deviceLoad[devUuid]
//	}
//}
//
//func containsString(slice []string, s string) bool {
//	for _, item := range slice {
//		if item == s {
//			return true
//		}
//	}
//	return false
//}
