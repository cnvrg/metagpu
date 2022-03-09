package gpumgr

import (
	"context"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/sharecfg"
	log "github.com/sirupsen/logrus"
	v1core "k8s.io/api/core/v1"
)

func (m *GpuMgr) discoverAnonymousProcesses() {
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
		for _, c := range p.Spec.Containers {
			for _, config := range cfg.Configs {
				resourceName := v1core.ResourceName(config.ResourceName)
				if quantity, ok := c.Resources.Limits[resourceName]; ok {
					m.GpuProcesses = append(m.GpuProcesses, &GpuProcess{
						PodId:             p.Name,
						PodNamespace:      p.Namespace,
						PodMetagpuRequest: quantity.Value(),
						ResourceName:      config.ResourceName,
						Nodename:          p.Spec.NodeName,
					})
				}
			}
		}
	}
}

func (m *GpuMgr) isProcessAnonymouse(podId string) bool {
	for _, p := range m.GpuProcesses {
		if p.PodId == podId {
			return false
		}
	}
	return true
}
