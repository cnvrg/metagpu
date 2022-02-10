package deviceplugin

import (
	"sort"
	"strings"
)

type DeviceUuid string

type DeviceLoad struct {
	FreeShares int
	Metagpus   []string
}

type DeviceAllocationMap struct {
	LoadMap             map[DeviceUuid]*DeviceLoad
	MetagpusAllocations []string
}

func NewDeviceLoadMap(realDeviceIds []string, availableMetagpus []string) *DeviceAllocationMap {
	sort.Strings(availableMetagpus)
	var deviceLoad = make(map[DeviceUuid]*DeviceLoad)
	var deviceToMetagpus = make(map[DeviceUuid][]string)
	for _, deviceId := range realDeviceIds {
		for _, availableDevId := range availableMetagpus {
			if strings.Contains(availableDevId, deviceId) {
				deviceToMetagpus[DeviceUuid(deviceId)] = append(deviceToMetagpus[DeviceUuid(deviceId)], availableDevId)
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
