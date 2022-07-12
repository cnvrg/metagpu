package podexec

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sClientConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
)

var copyCache *mgctlCopyCache

func GetK8sClient() (client.Client, error) {
	l := log.WithFields(log.Fields{"context": "getK8sClient"})
	rc := k8sClientConfig.GetConfigOrDie()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		log.Fatalf("error adding to scheme, err: %s ", err)
	}
	controllerClient, err := client.New(rc, client.Options{Scheme: scheme})
	if err != nil {
		l.Errorf("error creating new client, err: %s", err)
		return nil, err
	}

	return controllerClient, nil
}

func CopymgctlToContainer(containerName, podId, ns string) {

	pe, err := NewPodExec(containerName, podId, ns)
	if err != nil {
		log.Error(err)
		return
	}

	if pe.shouldCopyMgctl() {
		pe.copyMgctl()
		pe.makeMgctlExecutable()
		pe.getCopyCache().setCache(podId) // in memory cache to skip pod exec command
	}
}

func (e *podExec) shouldCopyMgctl() bool {
	if e.getCopyCache().isCached(e.podName) {
		return false
	}
	l := log.WithFields(log.Fields{"containerName": e.containerName, "podName": e.podName})
	output, err := e.RunCommand([]string{"/usr/bin/ls", "/usr/bin"})
	if err != nil {
		l.Error(err)
		return false
	}
	files := strings.Split(output, "\n")
	for _, fileName := range files {
		if fileName == "mgctl" {
			return false
		}
	}
	l.Info("injecting mgctl bin")
	return true
}

func (e *podExec) RunCommand(command []string) (output string, err error) {
	e.cmd = command
	e.stdout = new(bytes.Buffer)
	if err := e.exec(); err != nil {
		return "", err
	}
	return e.stdout.String(), nil
}

func (e *podExec) makeMgctlExecutable() {
	e.cmd = []string{"chmod", "+x", "/usr/bin/mgctl"}
	e.stdout = new(bytes.Buffer)
	if err := e.exec(); err != nil {
		log.WithFields(log.Fields{"containerName": e.containerName, "podName": e.podName}).Error(err)
	}
}

func (e *podExec) copyMgctl() {
	l := log.WithFields(log.Fields{"containerName": e.containerName, "podName": e.podName})
	var err error
	e.cmd = []string{"cp", "/dev/stdin", "/usr/bin/mgctl"}
	e.stdin, err = e.getmgctlBinFile()
	if err != nil {
		l.Error(err)
		return
	}
	if err := e.exec(); err != nil {
		l.Error(err)
		return
	}
}

func (e *podExec) getmgctlBinFile() (*os.File, error) {
	mgctlFile := viper.GetString("mgctlTar")
	if _, err := os.Stat(mgctlFile); err == nil {
		return os.Open(mgctlFile)
	} else {
		return nil, err
	}
}

func (e *podExec) getCopyCache() *mgctlCopyCache {
	if copyCache == nil {
		copyCache = NewMgctlCopyCache()
	}
	return copyCache
}

func (e *podExec) exec() error {
	rc := k8sClientConfig.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(rc)
	if err != nil {
		return err
	}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return err
	}
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(e.podName).
		Namespace(e.podNs).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Stdin:     e.stdin != nil,
			Stdout:    e.stdout != nil,
			Stderr:    true,
			TTY:       false,
			Container: e.containerName,
			Command:   e.cmd,
		}, runtime.NewParameterCodec(scheme))
	exec, err := remotecommand.NewSPDYExecutor(rc, "POST", req.URL())
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{Stdin: e.stdin, Stdout: e.stdout, Stderr: &stderr, Tty: false})
	if err != nil {
		return err
	}
	stdError := stderr.String()
	if stdError != "" {
		return errors.New(stdError)
	}
	return nil
}

func NewPodExec(containerName, podId, ns string) (*podExec, error) {
	return &podExec{podName: podId, podNs: ns, containerName: containerName}, nil
}
