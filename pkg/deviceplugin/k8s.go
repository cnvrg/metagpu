package deviceplugin

import (
	log "github.com/sirupsen/logrus"
	v1apps "k8s.io/api/apps/v1"
	v1batch "k8s.io/api/batch/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sClientConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
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
