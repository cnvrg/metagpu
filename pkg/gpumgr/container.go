package gpumgr

import (
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	v1core "k8s.io/api/core/v1"
	"strings"
)

type GpuContainer struct {
	ContainerId       string
	ContainerName     string
	PodId             string
	PodNamespace      string
	PodMetagpuRequest int64
	ResourceName      string
	Nodename          string
	Processes         []*GpuProcess
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

func NewGpuContainer(containerId, containerName, podId, ns, resourceName, nodename string, metagpuRequests int64) *GpuContainer {
	p := &GpuContainer{
		ContainerId:       containerId,
		PodId:             podId,
		ContainerName:     containerName,
		PodNamespace:      ns,
		PodMetagpuRequest: metagpuRequests,
		ResourceName:      resourceName,
		Nodename:          nodename,
	}
	if viper.GetBool("mgctlAutoInject") {
		podexec.CopymgctlToContainer(p.ContainerName, p.PodId, p.PodNamespace)
	}
	return p
}
