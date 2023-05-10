package v1

import (
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/gpumgr"
)

func listDeviceProcesses(podId string, gpuMgr *gpumgr.GpuMgr) (containers []*pb.GpuContainer) {

	for _, container := range gpuMgr.GetProcesses(podId) {
		var gpuProcesses []*pb.DeviceProcess

		for _, p := range container.Processes {
			gpuProcesses = append(gpuProcesses, &pb.DeviceProcess{
				Uuid:           p.DeviceUuid,
				Pid:            p.Pid,
				Memory:         p.GpuMemory,
				Cmdline:        p.GetShortCmdLine(),
				User:           p.User,
				ContainerId:    p.ContainerId,
				GpuUtilization: p.GpuUtilization,
			})
		}
		var gpuDevices []*pb.ContainerDevice
		for _, device := range container.Devices {
			gpuDevices = append(gpuDevices, &pb.ContainerDevice{
				Device: &pb.Device{
					Uuid:              device.GpuDevice.UUID,
					Index:             uint32(device.GpuDevice.Index),
					Shares:            uint32(device.GpuDevice.Shares),
					GpuUtilization:    device.GpuDevice.Utilization.Gpu,
					MemoryUtilization: device.GpuDevice.Utilization.Memory,
					MemoryTotal:       device.GpuDevice.Memory.Total,
					MemoryFree:        device.GpuDevice.Memory.Free,
					MemoryUsed:        device.GpuDevice.Memory.Used,
					MemoryShareSize:   device.GpuDevice.Memory.ShareSize,
					ResourceName:      device.GpuDevice.ResourceName,
					NodeName:          device.GpuDevice.Nodename,
				},
				AllocatedShares: device.AllocatedShares,
			})
		}
		containers = append(containers, &pb.GpuContainer{
			ContainerId:      container.ContainerId,
			ContainerName:    container.ContainerName,
			PodId:            container.PodId,
			PodNamespace:     container.PodNamespace,
			MetagpuRequests:  container.PodMetagpuRequest,
			MetagpuLimits:    container.PodMetagpuLimit,
			ResourceName:     container.ResourceName,
			NodeName:         container.Nodename,
			ContainerDevices: gpuDevices,
			DeviceProcesses:  gpuProcesses,
		})
	}
	return
}
