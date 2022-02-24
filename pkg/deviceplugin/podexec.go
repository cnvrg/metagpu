package deviceplugin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sClientConfig "sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
)

type podExec struct {
	pod           *corev1.Pod
	containerName string
	cmd           []string
	stdin         io.Reader
	stdout        *bytes.Buffer
}

func copymgctlToContainer(containerId string) {

	pe, err := getPodByContainerId(containerId)
	if err != nil {
		log.Error(err)
		return
	}
	// TODO: create in-memory cache with tall the pods that's already has mgctl
	if shouldCopyMgctl(pe) {
		copyMgctl(pe)
		makeMgctlExecutable(pe)
	}
}

func getPodByContainerId(containerId string) (podExec *podExec, err error) {
	c, err := GetK8sClient()
	if err != nil {
		return nil, err
	}

	pl := &corev1.PodList{}
	if err := c.List(context.Background(), pl, []client.ListOption{}...); err != nil {
		return nil, err
	}
	for _, pod := range pl.Items {
		for _, sc := range pod.Status.ContainerStatuses {
			if strings.Contains(sc.ContainerID, containerId) {
				return newPodExec(&pod, sc.Name), nil
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("pod with containerId %s not found", containerId))
}

func shouldCopyMgctl(pe *podExec) bool {
	pe.cmd = []string{"ls", "/usr/bin"}
	pe.stdout = new(bytes.Buffer)
	if err := pe.exec(); err != nil {
		log.Error(err)
		return false
	}
	files := strings.Split(pe.stdout.String(), "\n")
	for _, fileName := range files {
		if fileName == "mgctl" {
			return false
		}
	}
	log.Infof("mgctl not found in %s, injecting...", pe.containerName)
	return true
}

func makeMgctlExecutable(pe *podExec) {
	pe.cmd = []string{"chmod", "+x", "/usr/bin/mgctl"}
	pe.stdout = new(bytes.Buffer)
	if err := pe.exec(); err != nil {
		log.Error(err)
	}
}

func copyMgctl(pe *podExec) {
	var e error
	pe.cmd = []string{"cp", "/dev/stdin", "/usr/bin/mgctl"}
	pe.stdin, e = getmgctlBinFile()
	if e != nil {
		log.Error(e)
		return
	}
	if err := pe.exec(); err != nil {
		log.Error(err)
		return
	}
}

func getmgctlBinFile() (*os.File, error) {
	mgctlFile := viper.GetString("mgctlTar")
	if _, err := os.Stat(mgctlFile); err == nil {
		return os.Open(mgctlFile)
	} else {
		return nil, err
	}
}

func newPodExec(pod *corev1.Pod, container string) *podExec {
	return &podExec{pod: pod, containerName: container}
}

func (p *podExec) exec() error {
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
		Name(p.pod.Name).
		Namespace(p.pod.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Stdin:     p.stdin != nil,
			Stdout:    p.stdout != nil,
			Stderr:    true,
			TTY:       false,
			Container: p.containerName,
			Command:   p.cmd,
		}, runtime.NewParameterCodec(scheme))
	exec, err := remotecommand.NewSPDYExecutor(rc, "POST", req.URL())
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	err = exec.Stream(remotecommand.StreamOptions{Stdin: p.stdin, Stdout: p.stdout, Stderr: &stderr, Tty: false})
	if err != nil {
		return err
	}
	e := stderr.String()
	if e != "" {
		return errors.New(e)
	}
	return nil
}
