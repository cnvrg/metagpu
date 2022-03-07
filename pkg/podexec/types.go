package podexec

import (
	"bytes"
	"io"
	corev1 "k8s.io/api/core/v1"
)

type podExec struct {
	pod           *corev1.Pod
	containerName string
	cmd           []string
	stdin         io.Reader
	stdout        *bytes.Buffer
}
