package pkg

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"time"
)

type FractionalAcceleratorDevicePlugin struct {
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
	server := grpc.NewServer([]grpc.ServerOption{}...)
	socket := fmt.Sprintf("%scnvrg-fracacc.sock", pluginapi.DevicePluginPath)
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
