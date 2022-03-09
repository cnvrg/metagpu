package v1

import (
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/gpumgr"
)

func listDeviceProcesses(podId string, gpuMgr *gpumgr.GpuMgr) (devProc []*pb.DeviceProcess) {

	for _, process := range gpuMgr.GetProcesses(podId) {
		devProc = append(devProc, &pb.DeviceProcess{
			Pid:             process.Pid,
			Uuid:            process.DeviceUuid,
			Memory:          process.GpuMemory,
			Cmdline:         process.GetShortCmdLine(),
			User:            process.User,
			ContainerId:     process.ContainerId,
			PodName:         process.PodId,
			PodNamespace:    process.PodNamespace,
			MetagpuRequests: process.PodMetagpuRequest,
			GpuUtilization:  process.GpuUtilization,
			ResourceName:    process.ResourceName,
			NodeName:        process.Nodename,
		})
	}
	return
}
