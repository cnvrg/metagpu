package pkg

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"path"
	"time"
)

type FractionalAcceleratorDevicePlugin struct {
}

func (f *FractionalAcceleratorDevicePlugin) dial(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
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

func (f *FractionalAcceleratorDevicePlugin) Register() error {
	conn, err := f.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)
	req := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base("/var/lib/kubelet/device-plugins/fractor.sock"),
		ResourceName: "cnvrg.io/metagpu",
		Options:      &pluginapi.DevicePluginOptions{},
	}
	if _, err := client.Register(context.Background(), req); err != nil {
		return err
	}
	return nil
}

func (f *FractionalAcceleratorDevicePlugin) GetDevicePluginOptions(ctx context.Context, empty *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

func (f *FractionalAcceleratorDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {

	devs := []*pluginapi.Device{{ID: "foo-bar", Health: pluginapi.Healthy}}
	_ = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	for {
		time.Sleep(1)
		_ = s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
	}
}

func (f *FractionalAcceleratorDevicePlugin) GetPreferredAllocation(ctx context.Context, request *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

func (f *FractionalAcceleratorDevicePlugin) Allocate(ctx context.Context, request *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	return &pluginapi.AllocateResponse{}, nil
}

func (f *FractionalAcceleratorDevicePlugin) PreStartContainer(ctx context.Context, request *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (f *FractionalAcceleratorDevicePlugin) Serve() error {
	os.Remove("/var/lib/kubelet/device-plugins/fractor.sock")
	server := grpc.NewServer([]grpc.ServerOption{}...)
	socket := fmt.Sprintf("%sfractor.sock", pluginapi.DevicePluginPath)
	sock, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	pluginapi.RegisterDevicePluginServer(server, f)
	if err := server.Serve(sock); err != nil {
		return err
	}

	return nil
}
