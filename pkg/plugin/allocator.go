package plugin

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/nvmlutils"
	log "github.com/sirupsen/logrus"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func NewDeviceAllocation(allocationSize, totalShares int, availableDevIds []string) *DeviceAllocation {
	sort.Strings(availableDevIds)
	devAlloc := &DeviceAllocation{
		AvailableDevIds:   availableDevIds,
		AllocationSize:    allocationSize,
		TotalSharesPerGpu: totalShares,
	}
	devAlloc.PrintAvailableDevIds()
	devAlloc.InitLoadMap()
	devAlloc.SetAllocations()
	return devAlloc
}

func (a *DeviceAllocation) InitLoadMap() {
	a.LoadMap = make([]*DeviceLoad, nvmlutils.GetTotalDevices())
	// build a map of real device id to meta device id
	for _, deviceId := range a.MetaDeviceIdsToRealDeviceIds() {
		for _, availableDevId := range a.AvailableDevIds {
			if !strings.Contains(availableDevId, deviceId) {
				continue
			}
			devIdx := metaDeviceIdToDeviceIndex(availableDevId)
			if a.LoadMap[devIdx] == nil {
				a.LoadMap[devIdx] = &DeviceLoad{}
			}
			a.LoadMap[devIdx].Metagpus = append(a.LoadMap[devIdx].Metagpus, availableDevId)
		}
	}
}

func (a *DeviceAllocation) PrintAvailableDevIds() {
	log.Infof("[preferred-allocation] available (%d) devices IDs:", len(a.AvailableDevIds))
	for _, devId := range a.AvailableDevIds {
		log.Info(devId)
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
	entireGpusRequest := a.AllocationSize / a.TotalSharesPerGpu
	gpuFractionsRequest := a.AllocationSize % a.TotalSharesPerGpu
	log.Infof("metagpu allocation request: %d.%d", entireGpusRequest, gpuFractionsRequest)

	// first try to allocate entire gpus if requested
	for i := 0; i < entireGpusRequest; i++ {
		for _, devLoad := range a.LoadMap {
			if devLoad == nil {
				continue
			}
			if devLoad.getFreeShares() == a.TotalSharesPerGpu {
				a.MetagpusAllocations = append(a.MetagpusAllocations, devLoad.Metagpus...)
				devLoad.Metagpus = nil
			}
		}
	}

	if gpuFractionsRequest > 0 {
		for _, devLoad := range a.LoadMap {
			if devLoad == nil {
				continue
			}
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
		ExitMultiGpuFractionAlloc:
			if allocationsLeft > 0 {
				for _, devLoad := range a.LoadMap {
					if devLoad == nil {
						continue
					}
					for _, device := range devLoad.Metagpus {
						a.MetagpusAllocations = append(a.MetagpusAllocations, device)
						allocationsLeft--
						if allocationsLeft == 0 {
							goto ExitMultiGpuFractionAlloc
						}
					}
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

func metaDeviceIdToDeviceIndex(metaDeviceId string) (deviceIndex int) {
	r, _ := regexp.Compile("-\\d+-")
	s := strings.ReplaceAll(r.FindString(metaDeviceId), "-", "")
	idx, err := strconv.Atoi(s)
	if err != nil {
		log.Error("can't detect physical device ID from meta device id, err: %s", err)
	}
	return idx
}
