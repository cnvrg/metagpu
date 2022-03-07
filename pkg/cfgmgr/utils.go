package cfgmgr

import (
	"context"
	"fmt"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/podexec"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

func UpdatePersistentConfigs(metaGpus int32) {
	log.Info("updating persistent configs")
	c, err := podexec.GetK8sClient()
	if err != nil {
		log.Errorf("unable to write persistent configs, err: %s", err)
		return
	}
	name := types.NamespacedName{Namespace: "kube-system", Name: "metagpu-device-plugin-config"}
	metaGpuCm := &corev1.ConfigMap{}
	if err := c.Get(context.Background(), name, metaGpuCm); err != nil {
		log.Errorf("unable to write persistent configs, err: %s", err)
		return
	}

	var updatedConfigs []string
	for _, cl := range strings.Split(metaGpuCm.Data["config.yaml"], "\n") {
		if strings.Contains(cl, "metaGpus") {
			updatedConfigs = append(updatedConfigs, fmt.Sprintf("metaGpus: %d", metaGpus))
		} else {
			updatedConfigs = append(updatedConfigs, cl)
		}
	}
	metaGpuCm.Data["config.yaml"] = strings.Join(updatedConfigs, "\n")
	if err := c.Update(context.Background(), metaGpuCm); err != nil {
		log.Error(err)
	}
}
