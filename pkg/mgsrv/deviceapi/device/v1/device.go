package v1

import (
	"context"
	pb "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/gpumgr"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type DeviceService struct {
	pb.UnimplementedDeviceServiceServer
	gpuMgr *gpumgr.GpuMgr
	vl     string // visibility level
	cvl    string // container visibility level ID
	dvl    string // device visibility level ID
}

func (s *DeviceService) LoadContext(ctx context.Context) error {

	s.gpuMgr = ctx.Value("gpuMgr").(*gpumgr.GpuMgr)
	if s.gpuMgr == nil {
		log.Fatalf("gpuMgr instance not set in context")
	}
	s.vl = ctx.Value("visibilityLevel").(string)
	s.cvl = ctx.Value("containerVl").(string)
	s.dvl = ctx.Value("deviceVl").(string)
	// stop execution if visibility level is empty
	if s.vl == "" {
		return status.Errorf(codes.Aborted, "can't detect visibility level for request: %s", s.vl)
	}
	// stop executing if container or device visibility level is empty
	if s.cvl == "" || s.dvl == "" {
		return status.Error(codes.Aborted, "can't detect visibility levels")
	}
	return nil
}

func (s *DeviceService) GetGpuContainers(ctx context.Context, r *pb.GetGpuContainersRequest) (*pb.GetGpuContainersResponse, error) {

	if err := s.LoadContext(ctx); err != nil {
		return &pb.GetGpuContainersResponse{}, err
	}
	response := &pb.GetGpuContainersResponse{VisibilityLevel: s.vl}
	// stop execution if visibility level is container and pod id is not set (not enough permissions)
	if s.vl == s.cvl && r.PodId == "" {
		return response, status.Errorf(codes.PermissionDenied, "missing pod id and visibility level is to low (%s), can't proceed", s.vl)
	}
	if s.vl == s.dvl {
		r.PodId = "" // for deviceVisibilityLevel server should return all running process on all containers
	}
	response.GpuContainers = listDeviceProcesses(r.PodId, s.gpuMgr)
	return response, nil
}

func (s *DeviceService) StreamGpuContainers(r *pb.StreamGpuContainersRequest, stream pb.DeviceService_StreamGpuContainersServer) error {

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
		response := &pb.StreamGpuContainersResponse{VisibilityLevel: s.vl}
		response.GpuContainers = listDeviceProcesses(r.PodId, s.gpuMgr)
		if err := stream.Send(response); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

}

func (s *DeviceService) GetDevices(ctx context.Context, r *pb.GetDevicesRequest) (*pb.GetDevicesResponse, error) {
	response := &pb.GetDevicesResponse{}
	if err := s.LoadContext(ctx); err != nil {
		return response, err
	}
	response.Device = make(map[string]*pb.Device)
	for _, device := range s.gpuMgr.GetMetaDevices() {
		d := &pb.Device{
			Uuid:              device.UUID,
			Index:             uint32(device.Index),
			Shares:            uint32(device.Shares),
			GpuUtilization:    device.Utilization.Gpu,
			MemoryUtilization: device.Utilization.Memory,
			MemoryShareSize:   device.Memory.ShareSize,
			ResourceName:      device.ResourceName,
			NodeName:          device.Nodename,
		}
		if s.vl == s.dvl {
			d.MemoryTotal = device.Memory.Total
			d.MemoryFree = device.Memory.Free
			d.MemoryUsed = device.Memory.Used
		}
		response.Device[d.Uuid] = d
	}
	return response, nil
}

func (s *DeviceService) KillGpuProcess(ctx context.Context, r *pb.KillGpuProcessRequest) (*pb.KillGpuProcessResponse, error) {
	response := &pb.KillGpuProcessResponse{}
	if err := s.LoadContext(ctx); err != nil {
		return response, err
	}
	if err := s.gpuMgr.KillGpuProcess(r.Pid); err != nil {
		return response, status.Errorf(codes.Internal, "error killing GPU process, err: %s", err)
	}
	return response, nil
}

func (s *DeviceService) GetMetaDeviceInfo(ctx context.Context, r *pb.GetMetaDeviceInfoRequest) (*pb.GetMetaDeviceInfoResponse, error) {
	resp := &pb.GetMetaDeviceInfoResponse{}
	if err := s.LoadContext(ctx); err != nil {
		return resp, err
	}
	if s.vl != s.dvl {
		return resp, status.Errorf(codes.PermissionDenied, "wrong visibility level: %s", s.vl)
	}
	deviceInfo := s.gpuMgr.GetDeviceInfo()
	resp.Node = deviceInfo.Node
	resp.Metadata = deviceInfo.Metadata
	for _, device := range deviceInfo.Devices {
		resp.Devices = append(resp.Devices, &pb.Device{
			Uuid:              device.UUID,
			Index:             uint32(device.Index),
			Shares:            uint32(device.Shares),
			GpuUtilization:    device.Utilization.Gpu,
			MemoryUtilization: device.Utilization.Memory,
			MemoryShareSize:   device.Memory.ShareSize,
			MemoryTotal:       device.Memory.Total,
			MemoryFree:        device.Memory.Free,
			MemoryUsed:        device.Memory.Used,
			ResourceName:      device.ResourceName,
			NodeName:          device.Nodename,
		})
	}
	return resp, nil
}

func (s *DeviceService) PatchConfigs(ctx context.Context, r *pb.PatchConfigsRequest) (*pb.PatchConfigsResponse, error) {
	//if err := s.LoadContext(ctx); err != nil {
	//	return &pb.PatchConfigsResponse{}, err
	//}
	//if s.vl != s.dvl {
	//	return &pb.PatchConfigsResponse{}, status.Errorf(codes.PermissionDenied, "visibility level too high", s.vl)
	//}
	//deviceplugin.UpdatePersistentConfigs(r.MetagpusPerGpu)
	//viper.Set("metaGpus", r.MetagpusPerGpu)
	//s.gpuMgr.MetaGpuRecalculation <- true
	return &pb.PatchConfigsResponse{}, nil

}

func (s *DeviceService) PingServer(ctx context.Context, r *pb.PingServerRequest) (*pb.PingServerResponse, error) {
	return &pb.PingServerResponse{}, nil
}
