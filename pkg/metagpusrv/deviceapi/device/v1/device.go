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

	for _, device := range plugin.ListCachedDeviceProcesses() {
		var deviceProcesses []*pb.DeviceProcess
		response.Processes = map[string]*pb.DeviceProcesses{device.K8sDevice.ID: nil}
		for _, process := range device.Processes {
			deviceProcesses = append(deviceProcesses, &pb.DeviceProcess{
				Pid:         process.Pid,
				Memory:      process.Memory,
				Cmdline:     process.Cmdline,
				User:        process.User,
				ContainerId: process.ContainerId,
			})
		}
		response.Processes[device.K8sDevice.ID] = &pb.DeviceProcesses{DeviceProcess: deviceProcesses}
	}
	return response, nil
}
