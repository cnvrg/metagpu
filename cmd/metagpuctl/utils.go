package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
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
