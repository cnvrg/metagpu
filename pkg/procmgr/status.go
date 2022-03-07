package procmgr

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/utils"
	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

var MB uint64 = 1024 * 1024

type GpuStatus struct {
	SharingConfigs DevicesSharesConfigs
	GpuDevices     []*GpuDevice
	GpuProcesses   []*GpuProcess
}

func (s *GpuStatus) setGpuDevices() {
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
	s.GpuDevices = gpuDevices
}

func (s GpuStatus) setGpuProcesses() {
	var gpuProcesses []*GpuProcess
	for _, device := range s.GpuDevices {
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
	s.GpuProcesses = gpuProcesses
}
