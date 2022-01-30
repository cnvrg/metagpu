package pkg

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

var (
	UNIX_SOCKET  = "meta-fractor.sock"
	RESOUCE_NAME = "cnvrg.io/metagpu"
)

type MetaFractorDevicePlugin struct {
	server       *grpc.Server
	socket       string
	resourceName string
}

func (p *MetaFractorDevicePlugin) dial(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
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

func (p *MetaFractorDevicePlugin) Register() error {
	conn, err := p.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(p.socket),
		ResourceName: p.resourceName,
		Options:      &pluginapi.DevicePluginOptions{},
	}
	if _, err := client.Register(context.Background(), req); err != nil {
		return err
	}
	return nil
}

func (p *MetaFractorDevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (p *MetaFractorDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {

	log.Info("listAndWatch triggered...")
	devs := []*pluginapi.Device{
		{ID: "cnvrg-meta-device-0", Health: pluginapi.Healthy},
		{ID: "cnvrg-meta-device-1", Health: pluginapi.Healthy},
	}

	_ = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	for {
		select {

		//_ = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		}

	}
}

func (p *MetaFractorDevicePlugin) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (p *MetaFractorDevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	allocResponse := &pluginapi.AllocateResponse{}
	for _, req := range request.ContainerRequests {

		response := pluginapi.ContainerAllocateResponse{}
		//uuids := req.DevicesIDs
		response.Envs = map[string]string{
			"CNVRG_FOO":    "CNVRG_BAR",
			"CNVRG_DEVICE": strings.Join(req.DevicesIDs, ","),
		}
		allocResponse.ContainerResponses = append(allocResponse.ContainerResponses, &response)

	}
	return allocResponse, nil
}

func (p *MetaFractorDevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (p *MetaFractorDevicePlugin) Serve() error {
	_ = os.Remove(p.socket)

	sock, err := net.Listen("unix", p.socket)
	if err != nil {
		log.Error(err)
	}
	log.Infof("listening on %s", p.socket)
	pluginapi.RegisterDevicePluginServer(p.server, p)

	go func() {
		if err := p.server.Serve(sock); err != nil {
			log.Errorf("GRPC server craeshed, %s", err)
		}
	}()

	if conn, err := p.dial(p.socket, 3*time.Second); err != nil {
		log.Error(err)
		return err
	} else {
		_ = conn.Close()
		log.Info("GRPC successfully started and ready accept new connections")
	}
	return nil

}

func NewMetaFractorDevicePlugin() *MetaFractorDevicePlugin {
	return &MetaFractorDevicePlugin{
		server:       grpc.NewServer([]grpc.ServerOption{}...),
		socket:       fmt.Sprintf("%s%s", pluginapi.DevicePluginPath, UNIX_SOCKET),
		resourceName: RESOUCE_NAME,
	}
}
