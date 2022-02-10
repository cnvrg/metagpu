package v1

import (
	"context"
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
	log "github.com/sirupsen/logrus"
)

type DeviceService struct {
	pb.UnimplementedDeviceServiceServer
}

func (s *DeviceService) ListDeviceProcesses(ctx context.Context, r *pb.ListDeviceProcessesRequest) (*pb.ListDeviceProcessesResponse, error) {
	response := &pb.ListDeviceProcessesResponse{}
	plugin := ctx.Value("plugin").(*deviceplugin.MetaGpuDevicePlugin)
	if plugin == nil {
		log.Fatalf("plugin instance not set in context")
	}

	for deviceUuid, deviceProcesses := range plugin.ListDeviceProcesses() {
		for _, process := range deviceProcesses {

			response.DevicesProcesses = append(response.DevicesProcesses, &pb.DeviceProcess{
				Uuid:            string(deviceUuid),
				Pid:             process.Pid,
				Memory:          process.GpuMemory,
				Cmdline:         process.GetShortCmdLine(),
				User:            process.User,
				ContainerId:     process.ContainerId,
				PodName:         process.PodId,
				PodNamespace:    process.PodNamespace,
				MetagpuRequests: process.PodMetagpuRequest,
			})
		}
	}
	return response, nil
}
