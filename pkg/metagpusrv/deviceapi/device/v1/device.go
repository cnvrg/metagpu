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

func (s *DeviceService) ListProcesses(ctx context.Context, r *pb.ListProcessesRequest) (*pb.ListProcessesResponse, error) {

	if err := s.LoadContext(ctx); err != nil {
		return &pb.ListProcessesResponse{}, err
	}
	response := &pb.ListProcessesResponse{VisibilityLevel: s.vl}
	// stop execution if visibility level is container and pod id is not set (not enough permissions)
	if s.vl == s.cvl && r.PodId == "" {
		return response, status.Errorf(codes.PermissionDenied, "missing pod id and visibility level is to low (%s), can't proceed", s.vl)
	}
	if s.vl == s.dvl {
		r.PodId = "" // for deviceVisibilityLevel server should return all running process on all containers
	}
	response.DevicesProcesses = listDeviceProcesses(r.PodId, s.plugin)
	return response, nil
}

func (s *DeviceService) StreamProcesses(r *pb.StreamProcessesRequest, stream pb.DeviceService_StreamProcessesServer) error {

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
		response := &pb.StreamProcessesResponse{VisibilityLevel: s.vl}
		response.DevicesProcesses = listDeviceProcesses(r.PodId, s.plugin)
		if err := stream.Send(response); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

}

func (s *DeviceService) ListDevices(ctx context.Context, r *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	response := &pb.ListDevicesResponse{}
	if err := s.LoadContext(ctx); err != nil {
		return response, err
	}

	for _, device := range s.plugin.ListDevices() {
		response.Device = append(response.Device, &pb.Device{
			Uuid:              device.UUID,
			Index:             uint32(device.Index),
			Shares:            uint32(device.Shares),
			GpuUtilization:    device.Utilization.Gpu,
			MemoryUtilization: device.Utilization.Memory,
			MemoryTotal:       device.Memory.Total,
			MemoryFree:        device.Memory.Free,
			MemoryUsed:        device.Memory.Used,
			MemoryShareSize:   device.Memory.ShareSize,
		})
	}
	return response, nil
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
