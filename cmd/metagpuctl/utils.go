package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func GetGrpcMetaGpuSrvClientConn() (*grpc.ClientConn, error) {
	log.Infof("initiating GRPC connection to %s", viper.GetString("metagpu-server-addr"))
	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(viper.GetString("metagpu-server-addr"), opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func authenticatedContext() context.Context {
	ctx := context.Background()
	md := metadata.Pairs("Authorization", viper.GetString("token"))
	return metadata.NewOutgoingContext(ctx, md)
}
