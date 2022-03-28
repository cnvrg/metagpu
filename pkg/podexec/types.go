package podexec

import (
	"bytes"
	"io"
)

type podExec struct {
	podName       string
	podNs         string
	containerName string
	cmd           []string
	stdin         io.Reader
	stdout        *bytes.Buffer
}
