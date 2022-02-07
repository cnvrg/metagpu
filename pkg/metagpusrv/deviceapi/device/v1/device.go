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
		var pbDeviceProcesses []*pb.DeviceProcess
		response.Processes = map[string]*pb.DeviceProcesses{}

		for _, process := range deviceProcesses {
			pbDeviceProcesses = append(pbDeviceProcesses, &pb.DeviceProcess{
				Pid:         process.Pid,
				Memory:      process.GpuMemory,
				Cmdline:     process.GetShortCmdLine(),
				User:        process.User,
				ContainerId: process.ContainerId,
			})
		}

		response.Processes[deviceUuid] = &pb.DeviceProcesses{DeviceProcess: pbDeviceProcesses}
	}
	return response, nil
}
