package deviceplugin

import (
	"sort"
	"strings"
)

type DeviceLoad struct {
	FreeShares int
	Metagpus   []string
}

type DeviceAllocationMap struct {
	LoadMap             map[string]*DeviceLoad
	MetagpusAllocations []string
}

func NewDeviceLoadMap(realDeviceIds []string, availableMetagpus []string) *DeviceAllocationMap {
	sort.Strings(availableMetagpus)
	var deviceLoad = make(map[string]*DeviceLoad)
	var deviceToMetagpus = make(map[string][]string)
	for _, deviceId := range realDeviceIds {
		for _, availableDevId := range availableMetagpus {
			if strings.Contains(availableDevId, deviceId) {
				deviceToMetagpus[deviceId] = append(deviceToMetagpus[deviceId], availableDevId)
			}
		}
	}
	for devId, metagpus := range deviceToMetagpus {
		deviceLoad[devId] = &DeviceLoad{
			FreeShares: len(metagpus),
			Metagpus:   metagpus,
		}
	}
	return &DeviceAllocationMap{LoadMap: deviceLoad}
}

//func (m *DeviceAllocationMap) AllocateMetagpus(allocatableGPUs map[string]int) []string {
//	for devUuid, size := range allocatableGPUs {
//		for i = 0; i < deviceLoad[devUuid]
//	}
//}
