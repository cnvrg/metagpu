package v1

import (
	"context"
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/deviceplugin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type DeviceService struct {
	pb.UnimplementedDeviceServiceServer
	plugin *deviceplugin.MetaGpuDevicePlugin
	vl     string // visibility level
	cvl    string // container visibility level ID
	dvl    string // device visibility level ID
}

func (s *DeviceService) LoadContext(ctx context.Context) error {

	s.plugin = ctx.Value("plugin").(*deviceplugin.MetaGpuDevicePlugin)
	if s.plugin == nil {
		log.Fatalf("plugin instance not set in context")
	}
	s.vl = ctx.Value("visibilityLevel").(string)
	s.cvl = ctx.Value("containerVl").(string)
	s.dvl = ctx.Value("deviceVl").(string)
	// stop execution if visibility level is empty
	if s.vl == "" {
		return status.Errorf(codes.Aborted, "can't detect visibility level for request", s.vl)
	}
	// stop executing if container or device visibility level is empty
	if s.cvl == "" || s.dvl == "" {
		return status.Error(codes.Aborted, "can't detect visibility levels")
	}
	return nil
}

func (s *DeviceService) ListDeviceProcesses(ctx context.Context, r *pb.ListDeviceProcessesRequest) (*pb.ListDeviceProcessesResponse, error) {

	if err := s.LoadContext(ctx); err != nil {
		return &pb.ListDeviceProcessesResponse{}, err
	}
	response := &pb.ListDeviceProcessesResponse{VisibilityLevel: s.vl}
	// stop execution if visibility level is container and pod id is not set (not enough permissions)
	if s.vl == s.cvl && r.PodId == "" {
		return response, status.Errorf(codes.PermissionDenied, "missing pod id and visibility level is to low (%s), can't proceed", s.vl)
	}
	if s.vl == s.dvl {
		r.PodId = "" // for deviceVisibilityLevel server should return all running process on all containers
	}
	for deviceUuid, deviceProcesses := range s.plugin.ListDeviceProcesses(r.PodId) {
		for _, process := range deviceProcesses {
			if s.vl == s.cvl { // TODO: fix all this shit
				process.DeviceGpuUtilization = 0
				process.DeviceGpuMemoryUtilization = 0
				process.DeviceGpuMemoryTotal = 0
				process.DeviceGpuMemoryFree = 0
				process.TotalShares = 0
			}
			response.DevicesProcesses = append(response.DevicesProcesses, &pb.DeviceProcess{
				Uuid:                    string(deviceUuid),
				Pid:                     process.Pid,
				Memory:                  process.GpuMemory,
				Cmdline:                 process.GetShortCmdLine(),
				User:                    process.User,
				ContainerId:             process.ContainerId,
				PodName:                 process.PodId,
				PodNamespace:            process.PodNamespace,
				MetagpuRequests:         process.PodMetagpuRequest,
				DeviceGpuUtilization:    process.DeviceGpuUtilization,
				DeviceMemoryUtilization: process.DeviceGpuMemoryUtilization,
				DeviceMemoryTotal:       process.DeviceGpuMemoryTotal,
				DeviceMemoryFree:        process.DeviceGpuMemoryFree,
			})
		}
	}
	return response, nil
}

func (s *DeviceService) StreamDeviceProcesses(r *pb.StreamDeviceProcessesRequest, stream pb.DeviceService_StreamDeviceProcessesServer) error {

	for {

		if err := s.LoadContext(stream.Context()); err != nil {
			return err
		}
		// stop execution if visibility level is container and pod id is not set (not enough permissions)
		if s.vl == s.cvl && r.PodId == "" {
			return status.Errorf(codes.PermissionDenied, "missing pod id and visibility level is to low (%s), can't proceed", s.vl)
		}
		if s.vl == s.dvl {
			r.PodId = "" // for deviceVisibilityLevel server should return all running process on all containers
		}
		response := &pb.StreamDeviceProcessesResponse{VisibilityLevel: s.vl}
		for deviceUuid, deviceProcesses := range s.plugin.ListDeviceProcesses(r.PodId) {
			for _, process := range deviceProcesses {

				dp := &pb.DeviceProcess{
					Uuid:                    string(deviceUuid),
					Pid:                     process.Pid,
					Memory:                  process.GpuMemory,
					Cmdline:                 process.GetShortCmdLine(),
					User:                    process.User,
					ContainerId:             process.ContainerId,
					PodName:                 process.PodId,
					PodNamespace:            process.PodNamespace,
					MetagpuRequests:         process.PodMetagpuRequest,
					DeviceGpuUtilization:    process.DeviceGpuUtilization,
					DeviceMemoryUtilization: process.DeviceGpuMemoryUtilization,
					DeviceMemoryTotal:       process.DeviceGpuMemoryTotal,
					DeviceMemoryFree:        process.DeviceGpuMemoryFree,
					TotalShares:             int32(process.TotalShares),
				}
				if s.vl == s.cvl { // TODO: fix all this shit
					dp.DeviceGpuUtilization = 0
					dp.DeviceMemoryUtilization = 0
					dp.DeviceMemoryTotal = 0
					dp.DeviceMemoryFree = 0
					dp.TotalShares = 0
				}
				response.DevicesProcesses = append(response.DevicesProcesses, dp)
			}
		}
		if err := stream.Send(response); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

}

func (s *DeviceService) KillGpuProcess(ctx context.Context, r *pb.KillGpuProcessRequest) (*pb.KillGpuProcessResponse, error) {
	response := &pb.KillGpuProcessResponse{}
	if err := s.LoadContext(ctx); err != nil {
		return response, err
	}
	if err := s.plugin.KillGpuProcess(r.Pid); err != nil {
		return response, status.Errorf(codes.Internal, "error killing GPU process, err: %s", err)
	}
	return response, nil
}

func (s *DeviceService) PatchConfigs(ctx context.Context, r *pb.PatchConfigsRequest) (*pb.PatchConfigsResponse, error) {
	if err := s.LoadContext(ctx); err != nil {
		return &pb.PatchConfigsResponse{}, err
	}
	if s.vl != s.dvl {
		return &pb.PatchConfigsResponse{}, status.Errorf(codes.PermissionDenied, "visibility level to high", s.vl)
	}
	// override runtime configuration
	// TODO: make configuration persistent
	viper.Set("metaGpus", r.MetaGpus)
	s.plugin.MetaGpuRecalculation <- true
	return &pb.PatchConfigsResponse{}, nil

}

func (s *DeviceService) PingServer(ctx context.Context, r *pb.PingServerRequest) (*pb.PingServerResponse, error) {
	return &pb.PingServerResponse{}, nil
}
