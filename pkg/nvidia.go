package pkg

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"time"
)

type NvidiaDeviceManager struct {
	devices  []*pluginapi.Device
	cacheTTL time.Duration
}

func (m *NvidiaDeviceManager) CacheDevices() {
	m.setDevices()
	for {
		<-time.After(m.cacheTTL)
		m.setDevices()
	}
}

func (m *NvidiaDeviceManager) ListDevices() []*pluginapi.Device {
	return m.devices
}

func (m *NvidiaDeviceManager) setDevices() {

	log.Info("caching nvidia devices")
	count, ret := nvml.DeviceGetCount()
	nvmlErrorCheck(ret)
	var dl []*pluginapi.Device
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		nvmlErrorCheck(ret)
		uuid, ret := device.GetUUID()
		nvmlErrorCheck(ret)
		dl = append(dl, &pluginapi.Device{ID: uuid, Health: pluginapi.Healthy})
		log.Infof("discovered device: %s", uuid)
	}
	m.devices = dl
}

func NewNvidiaDeviceManager() *NvidiaDeviceManager {

	if ret := nvml.Init(); ret != nvml.SUCCESS {
		log.Fatalf(nvml.ErrorString(ret))
	}
	ndm := &NvidiaDeviceManager{cacheTTL: time.Second * 5}
	ndm.CacheDevices()
	return ndm
}

func nvmlErrorCheck(ret nvml.Return) {
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}
