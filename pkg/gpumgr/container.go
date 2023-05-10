package gpumgr

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1core "k8s.io/api/core/v1"
	"regexp"
	"strings"
)

type ContainerDevice struct {
	GpuDevice       *GpuDevice
	AllocatedShares int32
}

type GpuContainer struct {
	ContainerId       string
	ContainerName     string
	PodId             string
	PodNamespace      string
	PodMetagpuRequest int64
	PodMetagpuLimit	  int64
	ResourceName      string
	Nodename          string
	Processes         []*GpuProcess
	Devices           []*ContainerDevice
}

func getContainerId(pod *v1core.Pod, containerName string) (containerId string) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
			idx := strings.Index(status.ContainerID, "//")
			if idx != -1 {
				return status.ContainerID[idx+2:]
			} else {
				log.WithField("pod", pod.Name).Error("can't extract container id")
			}
		}
	}
	return
}

func (c *GpuContainer) setAllocatedGpus(gpuDevices []*GpuDevice) {
	l := log.WithField("pod", c.PodId)
	pe, err := podexec.NewPodExec(c.ContainerName, c.PodId, c.PodNamespace)
	if err != nil {
		l.Error(err)
		return
	}
	output, err := pe.RunCommand([]string{"printenv", "CNVRG_META_GPU_DEVICES"})
	if err != nil {
		l.Error(err)
		return
	}
	var gpuAllocationMap = make(map[string]int32)
	for _, metaDeviceId := range strings.Split(output, ",") {
		r, _ := regexp.Compile("cnvrg-meta-\\d+-\\d+-")
		deviceUuid := strings.TrimSuffix(r.ReplaceAllString(metaDeviceId, ""), "\n")
		if _, ok := gpuAllocationMap[deviceUuid]; ok {
			gpuAllocationMap[deviceUuid] = gpuAllocationMap[deviceUuid] + 1
		} else {
			gpuAllocationMap[deviceUuid] = 0
		}
	}

	for uuid, allocatedShares := range gpuAllocationMap {
		for _, device := range gpuDevices {
			if device.UUID == uuid {
				c.Devices = append(c.Devices, &ContainerDevice{
					GpuDevice:       device,
					AllocatedShares: allocatedShares,
				})
			}
		}
	}
}

func NewGpuContainer(containerId, containerName, podId, ns, resourceName, nodename string, metagpuRequests int64, metagpuLimits int64, gpuDevices []*GpuDevice) *GpuContainer {
	p := &GpuContainer{
		ContainerId:       containerId,
		PodId:             podId,
		ContainerName:     containerName,
		PodNamespace:      ns,
		PodMetagpuRequest: metagpuRequests,
		PodMetagpuLimit:   metagpuLimits,
		ResourceName:      resourceName,
		Nodename:          nodename,
	}
	// discover allocated GPUs
	p.setAllocatedGpus(gpuDevices)
	// inject mgctl bin
	if viper.GetBool("mgctlAutoInject") {
		podexec.CopymgctlToContainer(p.ContainerName, p.PodId, p.PodNamespace)
	}
	return p
}
