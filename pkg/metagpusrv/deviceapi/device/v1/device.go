package v1

import (
	"context"
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/metagpusrv"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	vl := ctx.Value("visibilityLevel").(string)
	if vl == "" {
		return response, status.Errorf(codes.Aborted, "can't detect visibility level for request", vl)
	}
	if metagpusrv.VisibilityLevel(vl) == metagpusrv.ContainerVisibility && r.PodId == "" {
		return response, status.Errorf(codes.Aborted, "missing pod id and visibility level is to low (%s) to proceed", vl)
	}
	if metagpusrv.VisibilityLevel(vl) == metagpusrv.DeviceVisibility {
		r.PodId = "" // for deviceVisibilityLevel server should return all running process on all containers
	}
	for deviceUuid, deviceProcesses := range plugin.ListDeviceProcesses(r.PodId) {
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
