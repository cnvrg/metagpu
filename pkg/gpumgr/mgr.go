package gpumgr

import (
	"context"
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/nvmlutils"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1core "k8s.io/api/core/v1"
	"os"
	"time"
)

var MB uint64 = 1024 * 1024

type GpuMgr struct {
	ContainerLevelVisibilityToken string
	DeviceLevelVisibilityToken    string
	GpuDevices                    []*GpuDevice
	//// list of gpu processes
	//gpuProcesses []*GpuProcess
	// list of gpu containers
	gpuContainers []*GpuContainer
	// collection of the gpu processes: the anonymouse and active running
	gpuContainersCollector []*GpuContainer
}

type GpuDeviceInfo struct {
	Node     string
	Metadata map[string]string
	Devices  []*GpuDevice
}

func (m *GpuMgr) startGpuStatusCache() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			// set gpu devices
			m.setGpuDevices()
			// set gpu containers
			m.discoverGpuContainers()
			// set active gpu processes
			m.enrichGpuContainer()
			// set final gpu containers list
			m.setGpuContainers()
		}
	}()
}

func (m *GpuMgr) setGpuDevices() {
	var gpuDevices []*GpuDevice
	for idx, device := range nvmlutils.GetDevices() {
		uuid, ret := device.GetUUID()
		nvmlutils.ErrorCheck(ret)
		deviceMemory, ret := device.GetMemoryInfo()
		nvmlutils.ErrorCheck(ret)
		utilization, ret := device.GetUtilizationRates()
		nvmlutils.ErrorCheck(ret)
		gpuDevices = append(gpuDevices, NewGpuDevice(uuid, idx, utilization, deviceMemory))
	}
	m.GpuDevices = gpuDevices
}

func (m *GpuMgr) enrichGpuContainer() {
	for _, device := range m.GpuDevices {
		for _, nvmlProcessInfo := range nvmlutils.GetComputeRunningProcesses(device.Index) {
			stats := nvmlutils.GetAccountingStats(device.Index, nvmlProcessInfo.Pid)
			gpuProc := NewGpuProcess(nvmlProcessInfo.Pid, stats.GpuUtilization, nvmlProcessInfo.UsedGpuMemory/MB, device.UUID)
			for _, c := range m.gpuContainersCollector {
				if c.ContainerId == gpuProc.ContainerId {
					c.Processes = append(c.Processes, gpuProc)
				}
			}
		}
	}
}

func (m *GpuMgr) setGpuContainers() {
	m.gpuContainers = m.gpuContainersCollector
	log.Infof("discovered %d gpu containers", len(m.gpuContainers))
}

func (m *GpuMgr) GetDeviceInfo() *GpuDeviceInfo {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to detect hostname, err: %s", err)
	}
	info := make(map[string]string)
	cudaVersion := nvmlutils.SystemGetCudaDriverVersion()
	info["cudaVersion"] = fmt.Sprintf("%d", cudaVersion)
	driver := nvmlutils.SystemGetDriverVersion()
	info["driverVersion"] = driver
	return &GpuDeviceInfo{Node: hostname, Metadata: info, Devices: m.GpuDevices}
}

func (m *GpuMgr) discoverGpuContainers() {
	c, err := podexec.GetK8sClient()
	if err != nil {
		log.Error(err)
		return
	}
	pl := &v1core.PodList{}
	if err := c.List(context.Background(), pl); err != nil {
		log.Error(err)
		return
	}
	// reset gpu containers collector
	m.gpuContainersCollector = nil
	cfg := sharecfg.NewDeviceSharingConfig()
	for _, p := range pl.Items {
		for _, container := range p.Spec.Containers {
			for _, config := range cfg.Configs {
				resourceName := v1core.ResourceName(config.ResourceName)
				if quantity, ok := container.Resources.Limits[resourceName]; ok {
					m.gpuContainersCollector = append(m.gpuContainersCollector,
						NewGpuContainer(
							getContainerId(&p, container.Name),
							container.Name,
							p.Name,
							p.Namespace,
							config.ResourceName,
							p.Spec.NodeName,
							quantity.Value(),
							m.GpuDevices,
						),
					)
				}
			}
		}
	}
}

func (m *GpuMgr) GetProcesses(podId string) []*GpuContainer {
	// if podId is set, return single process
	if podId != "" {
		var gpuContainers []*GpuContainer
		for _, deviceProcess := range m.gpuContainers {
			if deviceProcess.PodId == podId {
				gpuContainers = append(gpuContainers, deviceProcess)
			}
		}
		return gpuContainers
	}
	// return all processes
	return m.gpuContainers
}

func (m *GpuMgr) GetMetaDevices() map[string]*GpuDevice {
	var deviceMap = make(map[string]*GpuDevice)
	for _, d := range m.GpuDevices {
		deviceMap[d.UUID] = d
	}
	return deviceMap
}

func (m *GpuMgr) KillGpuProcess(pid uint32) error {
	p := NewGpuProcess(pid, 0, 0, "")
	return p.Kill()
}

func (m *GpuMgr) SetDeviceLevelVisibilityToken(token string) {
	m.DeviceLevelVisibilityToken = token
}

func (m *GpuMgr) SetContainerLevelVisibilityToken(token string) {
	m.ContainerLevelVisibilityToken = token
}

func NewGpuManager() *GpuMgr {
	mgr := &GpuMgr{}
	// init gpu devices
	mgr.setGpuDevices()
	// init gpu containers
	mgr.discoverGpuContainers()
	// init active gpu processes
	mgr.enrichGpuContainer()
	// set gpu processes
	mgr.setGpuContainers()
	// start gpu devices and processes cache
	mgr.startGpuStatusCache()
	// start mem enforcer
	if viper.GetBool("memoryEnforcer") {
		mgr.StartMemoryEnforcer()
	}
	return mgr
}
