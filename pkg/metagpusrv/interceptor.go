package metagpusrv

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"time"
)

type MetaGpuServerStream struct {
	grpc.ServerStream
	ctx context.Context
	//plugin          *deviceplugin.MetaGpuDevicePlugin
	//VisibilityToken string
	//DeviceVl        string
	//ContainerVl     string
}

//func (s *MetaGpuServerStream) Context() context.Context {
//	ctx := context.WithValue(context.Background(), TokenVisibilityClaimName, s.VisibilityToken)
//	ctx = context.WithValue(ctx, "containerVl", string(ContainerVisibility))
//	ctx = context.WithValue(ctx, "plugin", s.plugin)
//	return context.WithValue(ctx, "deviceVl", string(DeviceVisibility))
//}

func (s *MetaGpuServer) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &MetaGpuServerStream{ServerStream: ss}
		if !s.IsMethodPublic(info.FullMethod) {
			visibility, err := authorize(ss.Context())
			if err != nil {
				return err
			}
			ctx := context.WithValue(ss.Context(), TokenVisibilityClaimName, visibility)
			ctx = context.WithValue(ctx, "containerVl", string(ContainerVisibility))
			ctx = context.WithValue(ctx, "deviceVl", string(DeviceVisibility))
			ctx = context.WithValue(ctx, "plugin", s.plugin)

			//wrapper.VisibilityToken = visibility
			//wrapper.ContainerVl = string(ContainerVisibility)
			//wrapper.DeviceVl = string(DeviceVisibility)
		}
		return handler(srv, wrapper)
	}
}

func (s *MetaGpuServer) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		if !s.IsMethodPublic(info.FullMethod) {
			visibility, err := authorize(ctx)
			if err != nil {
				return nil, err
			}
			ctx = context.WithValue(ctx, TokenVisibilityClaimName, visibility)
			ctx = context.WithValue(ctx, "containerVl", string(ContainerVisibility))
			ctx = context.WithValue(ctx, "deviceVl", string(DeviceVisibility))
		}
		ctx = context.WithValue(ctx, "plugin", s.plugin)
		h, err := handler(ctx, req)
		log.Infof("[method: %s duration: %s]", info.FullMethod, time.Since(start))
		return h, err
	}
}
