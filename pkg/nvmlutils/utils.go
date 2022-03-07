package nvmlutils

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	log "github.com/sirupsen/logrus"
)

func ErrorCheck(ret nvml.Return) {
	if ret == nvml.ERROR_NOT_FOUND {
		log.Warnf("nvml error: ERROR_NOT_FOUND: [a query to find an object was unsuccessful]")
		return
	}
	if ret == nvml.ERROR_NOT_SUPPORTED {
		log.Warnf("nvml error: ERROR_NOT_SUPPORTED: [device doesn't support this feature]")
		return
	}
	if ret == nvml.ERROR_NO_PERMISSION {
		log.Warnf("nvml error: ERROR_NO_PERMISSION: [user doesn't have permission to perform this operation]")
		return
	}
	if ret != nvml.SUCCESS {
		log.Fatalf("fatal error during nvml operation: %s", nvml.ErrorString(ret))
	}
}
