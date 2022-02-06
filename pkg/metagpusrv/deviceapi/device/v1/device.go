package v1

import (
	"context"
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
)

type DeviceService struct {
	pb.UnimplementedDeviceServiceServer
}

func (s *DeviceService) ListDeviceProcesses(ctx context.Context, r *pb.ListDeviceProcessesRequest) (*pb.ListDeviceProcessesResponse, error) {
	response := &pb.ListDeviceProcessesResponse{}
	ndm := deviceplugin.NvidiaDeviceManager{}
	ndm.DiscoverDeviceProcesses()
	for _, device := range ndm.Devices {
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
