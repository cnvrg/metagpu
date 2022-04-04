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
		containers = append(containers, &pb.GpuContainer{
			ContainerId:     container.ContainerId,
			ContainerName:   container.ContainerName,
			PodId:           container.PodId,
			PodNamespace:    container.PodNamespace,
			MetagpuRequests: container.PodMetagpuRequest,
			ResourceName:    container.ResourceName,
			NodeName:        container.Nodename,
			DeviceProcesses: gpuProcesses,
		})

	}
	return
}
