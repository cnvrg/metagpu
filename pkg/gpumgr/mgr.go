package gpumgr

import (
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/nvmlutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
)

var MB uint64 = 1024 * 1024

type GpuMgr struct {
	ContainerLevelVisibilityToken string
	DeviceLevelVisibilityToken    string
	GpuDevices                    []*GpuDevice
	GpuProcesses                  []*GpuProcess
}

type GpuDeviceInfo struct {
	Node     string
	Metadata map[string]string
	Devices  []*GpuDevice
}

func (m *GpuMgr) startGpuStatusCache() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			m.setGpuDevices()
			m.setGpuProcesses()
			m.discoverAnonymousProcesses()
		}
	}()
}

func (m *GpuMgr) setGpuDevices() {
	var gpuDevices []*GpuDevice
	for idx, device := range nvmlutils.GetDevices() {
		uuid, ret := device.GetUUID()
		nvmlutils.ErrorCheck(ret)
		deviceMemory, ret := device.GetMemoryInfo()
		nvmlutils.ErrorCheck(ret)
		utilization, ret := device.GetUtilizationRates()
		nvmlutils.ErrorCheck(ret)
		gpuDevices = append(gpuDevices, NewGpuDevice(uuid, idx, utilization, deviceMemory))
	}
	m.GpuDevices = gpuDevices
}

func (m *GpuMgr) setGpuProcesses() {
	var gpuProcesses []*GpuProcess
	for _, device := range m.GpuDevices {
		for _, nvmlProcessInfo := range nvmlutils.GetComputeRunningProcesses(device.Index) {
			stats := nvmlutils.GetAccountingStats(device.Index, nvmlProcessInfo.Pid)
			gpuProc := NewGpuProcess(nvmlProcessInfo.Pid, stats.GpuUtilization, nvmlProcessInfo.UsedGpuMemory/MB, device.UUID)
			gpuProcesses = append(gpuProcesses, gpuProc)
		}
	}
	m.GpuProcesses = gpuProcesses
}

func (m *GpuMgr) GetDeviceInfo() *GpuDeviceInfo {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to detect hostname, err: %s", err)
	}
	info := make(map[string]string)
	cudaVersion := nvmlutils.SystemGetCudaDriverVersion()
	info["cudaVersion"] = fmt.Sprintf("%d", cudaVersion)
	driver := nvmlutils.SystemGetDriverVersion()
	info["driverVersion"] = driver
	return &GpuDeviceInfo{Node: hostname, Metadata: info, Devices: m.GpuDevices}
}

func (m *GpuMgr) GetProcesses(podId string) []*GpuProcess {
	if podId != "" {
		var podProcesses []*GpuProcess
		for _, deviceProcess := range m.GpuProcesses {
			if deviceProcess.PodId == podId {
				podProcesses = append(podProcesses, deviceProcess)
			}
		}
		return podProcesses
	}
	return m.GpuProcesses
}

func (m *GpuMgr) GetMetaDevices() map[string]*GpuDevice {
	var deviceMap = make(map[string]*GpuDevice)
	for _, d := range m.GpuDevices {
		deviceMap[d.UUID] = d
	}
	return deviceMap
}

func (m *GpuMgr) KillGpuProcess(pid uint32) error {
	p := NewGpuProcess(pid, 0, 0, "")
	return p.Kill()
}

func (m *GpuMgr) SetDeviceLevelVisibilityToken(token string) {
	m.DeviceLevelVisibilityToken = token
}

func (m *GpuMgr) SetContainerLevelVisibilityToken(token string) {
	m.ContainerLevelVisibilityToken = token
}

func NewGpuManager() *GpuMgr {
	mgr := &GpuMgr{}
	// init gpu devices
	mgr.setGpuDevices()
	// init gpu processes
	mgr.setGpuProcesses()
	// start gpu devices and processes cache
	mgr.startGpuStatusCache()
	// start mem enforcer
	if viper.GetBool("memoryEnforcer") {
		mgr.StartMemoryEnforcer()
	}
	return mgr
}
