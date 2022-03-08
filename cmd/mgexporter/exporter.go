package main

import (
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/ctlutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	conn         *grpc.ClientConn
	devicesCache map[string]*pbdevice.Device
	hostname     = ""

	deviceShares = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "shares",
		Help:      "total shares for single gpu unit",
	}, []string{"device_uuid", "device_index", "resource_name", "node_name"})

	deviceMemTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "memory_total",
		Help:      "total memory per device",
	}, []string{"device_uuid", "device_index", "resource_name", "node_name"})

	deviceMemFree = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "memory_free",
		Help:      "free memory per device",
	}, []string{"device_uuid", "device_index", "resource_name", "node_name"})

	deviceMemUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "memory_used",
		Help:      "used memory per device",
	}, []string{"device_uuid", "device_index", "resource_name", "node_name"})

	deviceMemShareSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "device",
		Name:      "memory_share_size",
		Help:      "metagpu memory share size",
	}, []string{"device_uuid", "device_index", "resource_name", "node_name"})

	deviceProcessGpuUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "gpu_utilization",
		Help:      "gpu process utilization in percentage",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMemoryUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "memory_usage",
		Help:      "process gpu-memory usage",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMetagpuRequests = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "metagpu_requests",
		Help:      "total metagpu requests in deployment spec",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMaxAllowedMetagpuGPUUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "max_allowed_metagpu_gpu_utilization",
		Help:      "max allowed metagpu gpu utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMetagpuCurrentGPUUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "metagpu_current_gpu_utilization",
		Help:      "current metagpu gpu utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMaxAllowedMetaGpuMemory = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "max_allowed_metagpu_memory",
		Help:      "max allowed metagpu memory usage",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMetagpuCurrentMemoryUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "metagpu_current_memory_utilization",
		Help:      "current metagpu memory utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})
)

func getGpuProcesses() []*pbdevice.DeviceProcess {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetProcessesRequest{}
	ctx := ctlutils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetProcesses(ctx, req)
	if err != nil {
		log.Error(err)
	}
	return resp.DevicesProcesses
}

func getGpuDevicesInfo() []*pbdevice.Device {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetMetaDeviceInfoRequest{}
	ctx := ctlutils.AuthenticatedContext(viper.GetString("token"))
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
	ctx := ctlutils.AuthenticatedContext(viper.GetString("token"))
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

func setDevicesMetrics() {
	// GPU device metrics
	for _, d := range getGpuDevicesInfo() {
		labels := []string{d.Uuid, fmt.Sprintf("%d", d.Index), d.ResourceName, hostname}
		deviceShares.WithLabelValues(labels...).Set(float64(d.Shares))
		deviceMemTotal.WithLabelValues(labels...).Set(float64(d.MemoryTotal))
		deviceMemFree.WithLabelValues(labels...).Set(float64(d.MemoryFree))
		deviceMemUsed.WithLabelValues(labels...).Set(float64(d.MemoryUsed))
		deviceMemShareSize.WithLabelValues(labels...).Set(float64(d.MemoryShareSize))
	}
}

func setHostname() {
	hn, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}
	hostname = hn
}

func setProcessesMetrics() {
	// GPU processes metrics
	for _, p := range getGpuProcesses() {
		//if _, ok := devicesCache[p.Uuid]; !ok {
		//	log.Warnf("process's device uuid: %s doesn not exists ", p.Uuid)
		//	continue
		//}
		labels := []string{
			p.Uuid, fmt.Sprintf("%d", p.Pid), p.Cmdline, p.User, p.PodName, p.PodNamespace, p.ResourceName, hostname}
		// process gpu utilization
		deviceProcessGpuUtilization.WithLabelValues(labels...).Set(float64(p.GpuUtilization))
		// process memory usage
		deviceProcessMemoryUsage.WithLabelValues(labels...).Set(float64(p.Memory))
		// metagpu requests
		deviceProcessMetagpuRequests.WithLabelValues(labels...).Set(float64(p.MetagpuRequests))
		// max allowed gpu utilization (by metagpus)
		maxMetaGpuUtilization := -1
		if devicesCache[p.Uuid] != nil {
			maxMetaGpuUtilization = int((100 / devicesCache[p.Uuid].Shares) * uint32(p.MetagpuRequests))
		}
		deviceProcessMaxAllowedMetagpuGPUUtilization.WithLabelValues(labels...).Set(float64(maxMetaGpuUtilization))
		// calculate gpu utilization relatively to the total metagpu requests
		metaGpuUtilization := -1
		if p.GpuUtilization > 0 {
			metaGpuUtilization = int(p.GpuUtilization*100) / maxMetaGpuUtilization
		}
		deviceProcessMetagpuCurrentGPUUtilization.WithLabelValues(labels...).Set(float64(metaGpuUtilization))
		// max allowed memory usage (by metagpus)
		maxMetaMemory := -1
		if devicesCache[p.Uuid] != nil {
			maxMetaMemory = int(uint64(p.MetagpuRequests) * devicesCache[p.Uuid].MemoryShareSize)
		}

		deviceProcessMaxAllowedMetaGpuMemory.WithLabelValues(labels...).Set(float64(maxMetaMemory))
		// calculate gpu memory utilization relatively to the total metagpu requests
		metaMemUtilization := -1
		if maxMetaMemory > 0 {
			metaMemUtilization = (int(p.Memory) * 100) / maxMetaMemory
		}
		deviceProcessMetagpuCurrentMemoryUtilization.WithLabelValues(labels...).Set(float64(metaMemUtilization))
	}
}

func recordMetrics() {
	go func() {
		for {
			conn = ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString("mgsrv"))
			if conn == nil {
				log.Fatal("connection is nil, can't continue")
				continue
			}
			// load devices cache
			setGpuDevicesCache()
			// set devices level metrics
			setDevicesMetrics()
			// set processes level metrics
			setProcessesMetrics()
			// close grcp connections
			conn.Close()
			// clear the cache
			clearGpuDevicesCache()
			time.Sleep(15 * time.Second)
		}
	}()
}

func startExporter() {

	log.Info("starting metagpu metrics exporter")
	prometheus.MustRegister(deviceShares)
	prometheus.MustRegister(deviceMemTotal)
	prometheus.MustRegister(deviceMemFree)
	prometheus.MustRegister(deviceMemUsed)
	prometheus.MustRegister(deviceMemShareSize)
	prometheus.MustRegister(deviceProcessGpuUtilization)
	prometheus.MustRegister(deviceProcessMemoryUsage)
	prometheus.MustRegister(deviceProcessMetagpuRequests)
	prometheus.MustRegister(deviceProcessMaxAllowedMetagpuGPUUtilization)
	prometheus.MustRegister(deviceProcessMetagpuCurrentGPUUtilization)
	prometheus.MustRegister(deviceProcessMaxAllowedMetaGpuMemory)
	prometheus.MustRegister(deviceProcessMetagpuCurrentMemoryUtilization)
	setHostname()
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
