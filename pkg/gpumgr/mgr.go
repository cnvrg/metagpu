package gpumgr

import (
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/utils"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
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
		time.Sleep(5 * time.Second)
		m.setGpuDevices()
		m.setGpuProcesses()
	}()
}

func (m *GpuMgr) setGpuDevices() {
	count, ret := nvml.DeviceGetCount()
	utils.NvmlErrorCheck(ret)
	var gpuDevices []*GpuDevice
	for i := 0; i < count; i++ {
		nvidiaDevice, ret := nvml.DeviceGetHandleByIndex(i)
		utils.NvmlErrorCheck(ret)
		uuid, ret := nvidiaDevice.GetUUID()
		utils.NvmlErrorCheck(ret)
		deviceMemory, ret := nvidiaDevice.GetMemoryInfo()
		utils.NvmlErrorCheck(ret)
		utilization, ret := nvidiaDevice.GetUtilizationRates()
		utils.NvmlErrorCheck(ret)
		gpuDevices = append(gpuDevices, NewGpuDevice(uuid, i, utilization, deviceMemory))
	}
	m.GpuDevices = gpuDevices
}

func (m *GpuMgr) setGpuProcesses() {
	var gpuProcesses []*GpuProcess
	for _, device := range m.GpuDevices {
		nvidiaDevice, ret := nvml.DeviceGetHandleByIndex(device.Index)
		utils.NvmlErrorCheck(ret)
		processes, ret := nvidiaDevice.GetComputeRunningProcesses()
		utils.NvmlErrorCheck(ret)
		for _, nvmlProcessInfo := range processes {
			stats, ret := nvidiaDevice.GetAccountingStats(nvmlProcessInfo.Pid)
			utils.NvmlErrorCheck(ret)
			gpuProc := NewGpuProcess(nvmlProcessInfo.Pid, stats.GpuUtilization, nvmlProcessInfo.UsedGpuMemory/MB, device.UUID)
			gpuProcesses = append(gpuProcesses, gpuProc)
		}
	}
	m.GpuProcesses = gpuProcesses
}

func (m *GpuMgr) GetDeviceInfo() *GpuDeviceInfo {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to detect hostname, err: %m", err)
	}
	info := make(map[string]string)
	cudaVersion, ret := nvml.SystemGetCudaDriverVersion()
	utils.NvmlErrorCheck(ret)
	info["cudaVersion"] = fmt.Sprintf("%d", cudaVersion)
	driver, ret := nvml.SystemGetDriverVersion()
	utils.NvmlErrorCheck(ret)
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
	status := &GpuMgr{}
	// init gpu devices
	status.setGpuDevices()
	// init gpu processes
	status.setGpuProcesses()
	// start gpu devices and processes cache
	status.startGpuStatusCache()
	return status
}
