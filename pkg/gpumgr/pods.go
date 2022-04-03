package gpumgr

import (
	"context"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	log "github.com/sirupsen/logrus"
	v1core "k8s.io/api/core/v1"
)

func (m *GpuMgr) discoverAnonymousGpuProcesses() {
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
	cfg := sharecfg.NewDeviceSharingConfig()
	for _, p := range pl.Items {
		if !m.isProcessAnonymouse(p.Name) {
			continue
		}
		for _, container := range p.Spec.Containers {
			for _, config := range cfg.Configs {
				resourceName := v1core.ResourceName(config.ResourceName)
				if quantity, ok := container.Resources.Limits[resourceName]; ok {
					m.gpuProcessCollector = append(m.gpuProcessCollector,
						NewGpuPod(container.Name, p.Name, p.Namespace, config.ResourceName, p.Spec.NodeName, quantity.Value()))
				}
			}
		}
	}
}

func (m *GpuMgr) isProcessAnonymouse(podId string) bool {
	for _, p := range m.gpuProcessCollector {
		if p.PodId == podId {
			return false
		}
	}
	return true
}
