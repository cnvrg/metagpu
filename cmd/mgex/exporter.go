package main

import (
	"errors"
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
	"time"
)

var (
	conn         *grpc.ClientConn
	devicesCache map[string]*pbdevice.Device

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

	deviceProcessAbsoluteGpuUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "absolute_gpu_utilization",
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
	}, []string{"pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMaxAllowedMetagpuGPUUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "max_allowed_metagpu_gpu_utilization",
		Help:      "max allowed metagpu gpu utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMetagpuRelativeGPUUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "metagpu_relative_gpu_utilization",
		Help:      "relative to metagpu request gpu utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMaxAllowedMetaGpuMemory = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "max_allowed_metagpu_memory",
		Help:      "max allowed metagpu memory usage",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})

	deviceProcessMetagpuRelativeMemoryUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "metagpu",
		Subsystem: "process",
		Name:      "metagpu_relative_memory_utilization",
		Help:      "relative to metagpus request memory utilization",
	}, []string{"uuid", "pid", "cmdline", "user", "pod_name", "pod_namespace", "resource_name", "node_name"})
)

func getGpuContainers() []*pbdevice.GpuContainer {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetGpuContainersRequest{}
	ctx := ctlutils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetGpuContainers(ctx, req)
	if err != nil {
		log.Error(err)
		return nil
	}
	return resp.GpuContainers
}

func getGpuDevicesInfo() []*pbdevice.Device {
	devices := pbdevice.NewDeviceServiceClient(conn)
	req := &pbdevice.GetMetaDeviceInfoRequest{}
	ctx := ctlutils.AuthenticatedContext(viper.GetString("token"))
	resp, err := devices.GetMetaDeviceInfo(ctx, req)
	if err != nil {
		log.Error(err)
		return nil
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
		return nil
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
		labels := []string{d.Uuid, fmt.Sprintf("%d", d.Index), d.ResourceName, d.NodeName}
		deviceShares.WithLabelValues(labels...).Set(float64(d.Shares))
		deviceMemTotal.WithLabelValues(labels...).Set(float64(d.MemoryTotal))
		deviceMemFree.WithLabelValues(labels...).Set(float64(d.MemoryFree))
		deviceMemUsed.WithLabelValues(labels...).Set(float64(d.MemoryUsed))
		deviceMemShareSize.WithLabelValues(labels...).Set(float64(d.MemoryShareSize))
	}
}

func resetProcessLevelMetrics() {
	deviceProcessAbsoluteGpuUtilization.Reset()
	deviceProcessMemoryUsage.Reset()
	deviceProcessMetagpuRequests.Reset()
	deviceProcessMaxAllowedMetagpuGPUUtilization.Reset()
	deviceProcessMetagpuRelativeGPUUtilization.Reset()
	deviceProcessMaxAllowedMetaGpuMemory.Reset()
	deviceProcessMetagpuRelativeMemoryUtilization.Reset()
}

func setProcessesMetrics() {
	// reset metrics
	resetProcessLevelMetrics()
	// GPU processes metrics
	for _, c := range getGpuContainers() {
		// metagpu requests
		deviceProcessMetagpuRequests.WithLabelValues(
			c.PodId, c.PodNamespace, c.ResourceName, c.NodeName).Set(float64(c.MetagpuRequests))
		for _, p := range c.DeviceProcesses {
			// set labels for device process level metrics
			labels := []string{
				p.Uuid, fmt.Sprintf("%d", p.Pid), p.Cmdline, p.User, c.PodId, c.PodNamespace, c.ResourceName, c.NodeName}
			// if pid is 0 => pod running without GPU process within
			if p.Pid == 0 {
				// absolute memory and gpu usage
				deviceProcessAbsoluteGpuUtilization.WithLabelValues(labels...).Set(0)
				deviceProcessMemoryUsage.WithLabelValues(labels...).Set(0)
				// max (relative to metagpus request) allowed gpu and memory utilization
				deviceProcessMaxAllowedMetagpuGPUUtilization.WithLabelValues(labels...).Set(0)
				deviceProcessMaxAllowedMetaGpuMemory.WithLabelValues(labels...).Set(0)
				// relative gpu and memory utilization
				deviceProcessMetagpuRelativeGPUUtilization.WithLabelValues(labels...).Set(0)
				deviceProcessMetagpuRelativeMemoryUtilization.WithLabelValues(labels...).Set(0)
			} else {
				// absolute memory and gpu usage
				deviceProcessAbsoluteGpuUtilization.WithLabelValues(labels...).Set(float64(p.GpuUtilization))
				deviceProcessMemoryUsage.WithLabelValues(labels...).Set(float64(p.Memory))
				// max (relative to metagpus request) allowed gpu and memory utilization
				deviceProcessMaxAllowedMetagpuGPUUtilization.WithLabelValues(labels...).Set(getMaxAllowedMetagpuGPUUtilization(c))
				deviceProcessMaxAllowedMetaGpuMemory.WithLabelValues(labels...).Set(getMaxAllowedMetaGpuMemory(c))
				// relative gpu and memory utilization
				deviceProcessMetagpuRelativeGPUUtilization.WithLabelValues(labels...).Set(getRelativeGPUUtilization(c, p))
				deviceProcessMetagpuRelativeMemoryUtilization.WithLabelValues(labels...).Set(getRelativeMemoryUtilization(c, p))
			}
		}
	}
}

func getMaxAllowedMetagpuGPUUtilization(c *pbdevice.GpuContainer) float64 {
	l := log.WithField("pod", c.PodId)
	d, err := getFirstContainerDevice(c)
	if err != nil {
		l.Error(err)
		return 0
	}
	return float64((100 / d.Device.Shares) * uint32(c.MetagpuRequests))
}

func getMaxAllowedMetaGpuMemory(c *pbdevice.GpuContainer) float64 {
	l := log.WithField("pod", c.PodId)
	d, err := getFirstContainerDevice(c)
	if err != nil {
		l.Error(err)
		return 0
	}
	return float64(uint64(c.MetagpuRequests) * d.Device.MemoryShareSize)

}

func getFirstContainerDevice(c *pbdevice.GpuContainer) (*pbdevice.ContainerDevice, error) {
	if len(c.ContainerDevices) == 0 {
		return nil, errors.New("no allocated gpus found")
	}
	return c.ContainerDevices[0], nil
}

func getRelativeGPUUtilization(c *pbdevice.GpuContainer, p *pbdevice.DeviceProcess) float64 {
	l := log.WithField("pod", c.PodId)
	d, err := getFirstContainerDevice(c)
	if err != nil {
		l.Error(err)
		return 0
	}
	maxMetaGpuUtilization := (100 / d.Device.Shares) * uint32(c.MetagpuRequests)
	metaGpuUtilization := 0
	if p.GpuUtilization > 0 && maxMetaGpuUtilization > 0 {
		metaGpuUtilization = int((p.GpuUtilization * 100) / maxMetaGpuUtilization)
	}
	return float64(metaGpuUtilization)
}

func getRelativeMemoryUtilization(c *pbdevice.GpuContainer, p *pbdevice.DeviceProcess) float64 {
	l := log.WithField("pod", c.PodId)
	d, err := getFirstContainerDevice(c)
	if err != nil {
		l.Error(err)
		return 0
	}
	maxMetaMemory := int(uint64(c.MetagpuRequests) * d.Device.MemoryShareSize)
	metaMemUtilization := 0
	if maxMetaMemory > 0 {
		metaMemUtilization = (int(p.Memory) * 100) / maxMetaMemory
	}
	return float64(metaMemUtilization)
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
	prometheus.MustRegister(deviceProcessAbsoluteGpuUtilization)
	prometheus.MustRegister(deviceProcessMemoryUsage)
	prometheus.MustRegister(deviceProcessMetagpuRequests)
	prometheus.MustRegister(deviceProcessMaxAllowedMetagpuGPUUtilization)
	prometheus.MustRegister(deviceProcessMetagpuRelativeGPUUtilization)
	prometheus.MustRegister(deviceProcessMaxAllowedMetaGpuMemory)
	prometheus.MustRegister(deviceProcessMetagpuRelativeMemoryUtilization)
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
