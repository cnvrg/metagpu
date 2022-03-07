package mgsrv

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"time"
)

type MetaGpuServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *MetaGpuServerStream) Context() context.Context {
	return s.ctx
}

func (s *MetaGpuServer) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapper := &MetaGpuServerStream{ServerStream: ss}
		if !s.IsMethodPublic(info.FullMethod) {
			visibility, err := authorize(ss.Context())
			if err != nil {
				return err
			}
			wrapper.ctx = context.WithValue(ss.Context(), TokenVisibilityClaimName, visibility)
			wrapper.ctx = context.WithValue(wrapper.ctx, "containerVl", string(ContainerVisibility))
			wrapper.ctx = context.WithValue(wrapper.ctx, "deviceVl", string(DeviceVisibility))
			wrapper.ctx = context.WithValue(wrapper.ctx, "gpuMgr", s.gpuMgr)

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
		ctx = context.WithValue(ctx, "gpuMgr", s.gpuMgr)
		h, err := handler(ctx, req)
		if viper.GetBool("verbose") {
			log.Infof("[method: %s duration: %s]", info.FullMethod, time.Since(start))
		}
		return h, err
	}
}
