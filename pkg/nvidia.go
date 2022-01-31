package pkg

import (
	log "github.com/sirupsen/logrus"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type NvidiaDeviceManager struct{}

func (m *NvidiaDeviceManager) ListDevices() []*pluginapi.Device {
	log.Info("listing nvidia devices")
	return nil
}
