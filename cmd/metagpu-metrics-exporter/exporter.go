package main

import (
	"fmt"
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
	conn           *grpc.ClientConn
	devicesCache   map[string]*pbdevice.Device
	devicesOverall = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "overall",
		Help:      "metagpu devicesOverall info",
	}, []string{
		"device_uuid",
		"device_index",
		"shares",
		"memory_total",
		"memory_free",
		"memory_used",
		"memory_share_size",
	})
	processesOverall = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "overall",
		Help:      "metagpu process info",
	}, []string{
		"uuid",
		"pid",
		"gpu",
		"memory",
		"cmdline",
		"user",
		"pod_name",
		"pod_namespace",
		"metagpu_requests",
		"max_meta_gpu",
		"meta_gpu_utilization",
		"max_meta_memory",
		"meta_memory_utilization",
	})
)

func getGpuProcesses() []*pbdevice.DeviceProcess {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetProcessesRequest{}
	ctx := utils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetProcesses(ctx, req)
	if err != nil {
		log.Error(err)
	}
	return resp.DevicesProcesses
}

func getGpuDevicesInfo() []*pbdevice.Device {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetMetaDeviceInfoRequest{}
	ctx := utils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetMetaDeviceInfo(ctx, req)
	if err != nil {
		log.Error(err)
	}
	return resp.Devices
}

func setGpuDevicesCache() map[string]*pbdevice.Device {
	if devicesCache != nil {
		return devicesCache
	}
	devicesCache = make(map[string]*pbdevice.Device)
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetDevicesRequest{}
	ctx := utils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetDevices(ctx, req)
	if err != nil {
		log.Error(err)
	}
	devicesCache = resp.Device
	return devicesCache
}

func clearGpuDevicesCache() {
	devicesCache = nil
}

func recordMetrics() {
	go func() {
		for {
			conn = utils.GetGrpcMetaGpuSrvClientConn(viper.GetString("mgsrv"))
			if conn == nil {
				log.Fatal("connection is nil, can't continue")
				continue
			}
			setGpuDevicesCache()
			devicesOverall.Reset()
			processesOverall.Reset()
			for _, d := range getGpuDevicesInfo() {
				devicesOverall.WithLabelValues(
					d.Uuid,
					fmt.Sprintf("%d", d.Index),
					fmt.Sprintf("%d", d.Shares),
					fmt.Sprintf("%d", d.MemoryTotal),
					fmt.Sprintf("%d", d.MemoryFree),
					fmt.Sprintf("%d", d.MemoryUsed),
					fmt.Sprintf("%d", d.MemoryShareSize),
				).Set(1)
			}
			for _, p := range getGpuProcesses() {
				if _, ok := devicesCache[p.Uuid]; !ok {
					log.Warnf("process's device uuid: %s doesn not exists ", p.Uuid)
					continue
				}
				maxMetaGpu := (100 / devicesCache[p.Uuid].Shares) * uint32(p.MetagpuRequests)
				maxMetaMemory := uint64(p.MetagpuRequests) * devicesCache[p.Uuid].MemoryShareSize
				metaGpuUtilization := (p.GpuUtilization * 100) / maxMetaGpu
				metaMemUtilization := (p.Memory * 100) / maxMetaMemory
				processesOverall.WithLabelValues(
					p.Uuid,
					fmt.Sprintf("%d", p.Pid),
					fmt.Sprintf("%d", p.GpuUtilization),
					fmt.Sprintf("%d", p.Memory),
					p.Cmdline,
					p.User,
					p.PodName,
					p.PodNamespace,
					fmt.Sprintf("%d", p.MetagpuRequests),
					fmt.Sprintf("%d", maxMetaGpu),
					fmt.Sprintf("%d", metaGpuUtilization),
					fmt.Sprintf("%d", maxMetaMemory),
					fmt.Sprintf("%d", metaMemUtilization),
				)
			}
			conn.Close()
			clearGpuDevicesCache()
			time.Sleep(1 * time.Second)
		}
	}()
}

func startExporter() {

	log.Info("starting metagpu metrics exporter")
	prometheus.MustRegister(devicesOverall)
	recordMetrics()
	addr := viper.GetString("metrics-addr")
	http.Handle("/metrics", promhttp.Handler())
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error(err)
		return
	}
	log.Infof("metrics serving on http://%s/metrics", addr)
	if err := http.Serve(l, nil); err != nil {
		log.Error(err)
		return
	}
}
