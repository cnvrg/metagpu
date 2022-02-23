package deviceplugin

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	v1apps "k8s.io/api/apps/v1"
	v1batch "k8s.io/api/batch/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sClientConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
)

func GetK8sClient() (client.Client, error) {
	l := log.WithFields(log.Fields{"context": "getK8sClient"})
	rc := k8sClientConfig.GetConfigOrDie()
	scheme := runtime.NewScheme()
	if err := v1core.AddToScheme(scheme); err != nil {
		l.Fatalf("error adding to scheme, err: %s ", err)
	}
	if err := v1apps.AddToScheme(scheme); err != nil {
		l.Fatalf("error adding to scheme, err: %s ", err)
	}
	if err := v1batch.AddToScheme(scheme); err != nil {
		l.Fatalf("error adding to scheme, err: %s ", err)
	}

	controllerClient, err := client.New(rc, client.Options{Scheme: scheme})
	if err != nil {
		l.Errorf("error creating new client, err: %s", err)
		return nil, err
	}

	return controllerClient, nil
}

func GetRestConfigs() (*rest.Config, error) {
	return k8sClientConfig.GetConfigOrDie(), nil

}

func UpdatePersistentConfigs(metaGpus int32) {
	log.Info("updating persistent configs")
	c, err := GetK8sClient()
	if err != nil {
		log.Error(err)
		return
	}
	name := types.NamespacedName{Namespace: "kube-system", Name: "metagpu-device-plugin-config"}
	metaGpuCm := &v1core.ConfigMap{}
	if err := c.Get(context.Background(), name, metaGpuCm); err != nil {
		log.Error(err)
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
