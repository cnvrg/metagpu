package podexec

import (
	"bytes"
	"io"
	"sync"
)

type mgctlCopyCache struct {
	mu    sync.Mutex
	cache map[string]bool
}

type podExec struct {
	podName       string
	podNs         string
	containerName string
	cmd           []string
	stdin         io.Reader
	stdout        *bytes.Buffer
}
