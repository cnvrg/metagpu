package deviceplugin

import (
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
	MetagpusAllocations []string
}

func NewDeviceAllocation(allocationSize int, availableDevIds []string) *DeviceAllocation {
	sort.Strings(availableDevIds)
	devAlloc := &DeviceAllocation{AllocationSize: allocationSize, AvailableDevIds: availableDevIds}
	devAlloc.InitLoadMap()
	return devAlloc
}

func (a *DeviceAllocation) InitLoadMap() {

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
