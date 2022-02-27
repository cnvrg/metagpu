package main

import (
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
)

var (
	conn    *grpc.ClientConn
	devices = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "devices",
		Name:      "overall",
		Help:      "metagpu devices info",
	}, []string{
		"device_uuid",
		"device_index",
		"shares",
		"memory_total",
		"memory_free",
		"memory_used",
		"memory_share_size",
	})
)

func getGpuDevices() {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetMetaDeviceInfoRequest{}
	ctx := utils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetMetaDeviceInfo(ctx, req)
	if err != nil {
		log.Error(err)
	}
	log.Info(resp)
}

func recordMetrics() {
	go func() {
		for {
			getGpuDevices()
			time.Sleep(1 * time.Second)
		}
	}()
}

func startExporter() {
	conn = utils.GetGrpcMetaGpuSrvClientConn(viper.GetString("mgsrv"))
	if conn == nil {
		log.Fatal("connection is nil, can't continue")
	}
	log.Info("starting metagpu metrics exporter")
	recordMetrics()
	http.Handle("/metrics", promhttp.Handler())
	l, err := net.Listen("tcp", viper.GetString("api.metricsAddr"))
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("metrics serving on http://%s/metrics", viper.GetString("metrics-addr"))
	if err := http.Serve(l, nil); err != nil {
		log.Error(err)
		return
	}
}
