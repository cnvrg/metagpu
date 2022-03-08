package nvmlutils

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
)

func init() {
	ret := nvml.Init()
	ErrorCheck(ret)
}

func GetDevices() (devices []*nvml.Device) {
	count, ret := nvml.DeviceGetCount()
	ErrorCheck(ret)
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		ErrorCheck(ret)
		devices = append(devices, &device)
	}
	return
}

func GetComputeRunningProcesses(deviceIdx int) []nvml.ProcessInfo {
	processes, ret := getDeviceByIdx(deviceIdx).GetComputeRunningProcesses()
	ErrorCheck(ret)
	return processes
}

func GetAccountingStats(deviceIdx int, pid uint32) *nvml.AccountingStats {
	stats, ret := getDeviceByIdx(deviceIdx).GetAccountingStats(pid)
	ErrorCheck(ret)
	return &stats
}

func SystemGetCudaDriverVersion() int {
	cudaVersion, ret := nvml.SystemGetCudaDriverVersion()
	ErrorCheck(ret)
	return cudaVersion
}

func SystemGetDriverVersion() string {
	driver, ret := nvml.SystemGetDriverVersion()
	ErrorCheck(ret)
	return driver
}

func GetDeviceMemory(device *nvml.Device) *nvml.Memory {
	memInfo, ret := device.GetMemoryInfo()
	ErrorCheck(ret)
	return &memInfo
}

func GetDeviceByUUID(uuid string) *nvml.Device {
	for _, device := range GetDevices() {
		devUuid, ret := device.GetUUID()
		ErrorCheck(ret)
		if devUuid == uuid {
			return device
		}
	}
	return nil
}

func GetDeviceUUID(device *nvml.Device) string {
	uuid, ret := device.GetUUID()
	ErrorCheck(ret)
	return uuid
}

func ErrorCheck(ret nvml.Return) {
	if ret == nvml.ERROR_NOT_FOUND {
		log.Warnf("nvml error: ERROR_NOT_FOUND: [a query to find an object was unsuccessful]")
		return
	}
	if ret == nvml.ERROR_NOT_SUPPORTED {
		log.Warnf("nvml error: ERROR_NOT_SUPPORTED: [device doesn't support this feature]")
		return
	}
	if ret == nvml.ERROR_NO_PERMISSION {
		log.Warnf("nvml error: ERROR_NO_PERMISSION: [user doesn't have permission to perform this operation]")
		return
	}
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}

func getDeviceByIdx(deviceIdx int) *nvml.Device {
	device, ret := nvml.DeviceGetHandleByIndex(deviceIdx)
	ErrorCheck(ret)
	return &device
}
