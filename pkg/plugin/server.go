package plugin

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

func (p *MetaGpuDevicePlugin) dial(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(socket, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return net.DialTimeout("unix", socket, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil

}

func (p *MetaGpuDevicePlugin) Register() error {
	conn, err := p.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(p.socket),
		ResourceName: p.GetDeviceSharingConfig().ResourceName,
		Options: &pluginapi.DevicePluginOptions{
			GetPreferredAllocationAvailable: true,
		},
	}
	if _, err := client.Register(context.Background(), req); err != nil {
		return err
	}
	return nil
}

func (p *MetaGpuDevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{GetPreferredAllocationAvailable: true}, nil
}

func (p *MetaGpuDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {

	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: p.GetPluginDevices()}); err != nil {
		log.Error(err)
	}

	for {
		select {
		case <-p.stop:
			return nil
		case <-p.MetaGpuRecalculation:
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: p.GetPluginDevices()}); err != nil {
				log.Error(err)
			}
		}
	}
}

func (p *MetaGpuDevicePlugin) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {

	allocResponse := &pluginapi.PreferredAllocationResponse{}
	for _, req := range request.ContainerRequests {
		allocContainerResponse := &pluginapi.ContainerPreferredAllocationResponse{}
		allocContainerResponse.DeviceIDs, _ = p.MetagpuAllocation(int(req.AllocationSize), req.GetAvailableDeviceIDs())
		log.Info("preferred devices ids:")
		for _, devId := range allocContainerResponse.DeviceIDs {
			log.Info(devId)
		}
		allocResponse.ContainerResponses = append(allocResponse.ContainerResponses, allocContainerResponse)
	}
	return allocResponse, nil

}

func (p *MetaGpuDevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	allocResponse := &pluginapi.AllocateResponse{}
	for _, req := range request.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{}
		sort.Strings(req.DevicesIDs)
		log.Info("requested devices ids:")
		for _, dev := range req.DevicesIDs {
			log.Info(dev)
		}
		metaGpuMaxMem := "" // TODO: fix this
		realDevices := p.ParseRealDeviceId(req.DevicesIDs)
		//if len(realDevices) > 0 {
		//	metaGpuMaxMem = fmt.Sprintf("%d", p.shareCfg.MetaGpus*(len(req.DevicesIDs)))
		//}
		response.Envs = map[string]string{
			"CNVRG_META_GPU_DEVICES": strings.Join(req.DevicesIDs, ","),
			"NVIDIA_VISIBLE_DEVICES": strings.Join(realDevices, ","),
			"METAGPU_MAX_MEM":        metaGpuMaxMem,
			"MG_CTL_ADDR":            fmt.Sprintf("%s:50052", os.Getenv("POD_IP")),
			"MG_CTL_TOKEN":           viper.GetString("containerToken"),
		}
		allocResponse.ContainerResponses = append(allocResponse.ContainerResponses, &response)
	}
	return allocResponse, nil
}

func (p *MetaGpuDevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *MetaGpuDevicePlugin) Serve() error {
	_ = os.Remove(p.socket)

	sock, err := net.Listen("unix", p.socket)
	if err != nil {
		log.Error(err)
	}
	log.Infof("listening on %s", p.socket)
	pluginapi.RegisterDevicePluginServer(p.server, p)

	go func() {
		if err := p.server.Serve(sock); err != nil {
			log.Errorf("gRPC server craeshed, %s", err)
		}
	}()

	if conn, err := p.dial(p.socket, 3*time.Second); err != nil {
		log.Error(err)
		return err
	} else {
		_ = conn.Close()
		log.Info("gRPC server successfully started and ready accept new connections")
	}
	return nil

}

func (p *MetaGpuDevicePlugin) Start() {
	if err := p.Serve(); err != nil {
		log.Fatal(err)
	}

	if err := p.Register(); err != nil {
		log.Fatal(err)
	}

}

func (p *MetaGpuDevicePlugin) Stop() {
	log.Info("stopping GRPC server")
	if p != nil && p.server != nil {
		p.server.Stop()
	}
	log.Info("removing unix socket")
	_ = os.Remove(p.socket)
	log.Info("closing all channels")
	close(p.stop)
	close(p.MetaGpuRecalculation)
}

func NewMetaGpuDevicePlugin(metaGpuRecalculation chan bool, deviceMgr DeviceManager) *MetaGpuDevicePlugin {
	if viper.GetString("accelerator") != "nvidia" {
		log.Fatal("accelerator not supported, currently only nvidia is supported")
	}
	return &MetaGpuDevicePlugin{
		server:               grpc.NewServer([]grpc.ServerOption{}...),
		socket:               fmt.Sprintf("%s%s", pluginapi.DevicePluginPath, deviceMgr.GetUnixSocket()),
		DeviceManager:        deviceMgr,
		stop:                 make(chan interface{}),
		MetaGpuRecalculation: metaGpuRecalculation,
	}
}
