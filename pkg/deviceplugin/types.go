package deviceplugin

import (
	"google.golang.org/grpc"
	"time"
)

type MetaGpuDevicePlugin struct {
	DeviceManager
	server       *grpc.Server
	socket       string
	resourceName string
	deviceUuids  []string
	totalShares  int
	//containerLevelVisibilityToken string
	//deviceLevelVisibilityToken    string
	stop                 chan interface{}
	MetaGpuRecalculation chan bool
}

type NvidiaDeviceManager struct {
	Devices []*MetaDevice
	//Processes                []*DeviceProcess
	cacheTTL                 time.Duration
	processesDiscoveryPeriod time.Duration
}

//type DeviceMemory struct {
//	Total     uint64
//	Free      uint64
//	Used      uint64
//	ShareSize uint64
//}
//
//type DeviceUtilization struct {
//	Gpu    uint32
//	Memory uint32
//}

type MetaDevice struct {
	UUID  string
	Index int
	//Shares      int
	//Utilization *DeviceUtilization
	//Memory *DeviceMemory
}

type DeviceUuid string

type DeviceLoad struct {
	Metagpus []string
}

type DeviceAllocation struct {
	LoadMap             []*DeviceLoad
	AvailableDevIds     []string
	AllocationSize      int
	TotalShares         int
	MetagpusAllocations []string
}
