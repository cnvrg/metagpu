package deviceplugin

import (
	"context"
	"errors"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"io"
	"strings"

	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"os/exec"
)

func copymgctlToContainer(containerId string) {
	if viper.GetString("cri") == "docker" {
		copymgctlToDockerContainer(containerId)
	}
	if viper.GetString("cri") == "containerd" {
		copymgctlToContainerdContainer(containerId)
	}
}

func copymgctlToDockerContainer(containerId string) {
	ctx := context.Background()
	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	defer cli.Close()
	if err != nil {
		log.Error(err)
		return
	}
	if !isDockerContainerExists(cli, containerId) {
		log.Errorf("container with ID: %s does not exists, can't copy mgctl binary", containerId)
		return
	}

	if !shouldCopyDocker(cli, containerId) {
		return
	}

	if f, err := getmgctlBinFile(); err == nil {
		if err := cli.CopyToContainer(ctx, containerId, "/usr/bin", f, types.CopyToContainerOptions{}); err != nil {
			log.Error(err)
		}
		_ = f.Close()
	} else {
		log.Error(err)
	}
}

func shouldCopyDocker(dc *docker.Client, containerId string) bool {
	_, err := dc.ContainerStatPath(context.Background(), containerId, "/usr/bin/mgctl")
	if err != nil {
		log.Warnf("mgctl not found, err: %s", err)
		return true
	}
	return false
}

func isDockerContainerExists(dc *docker.Client, containerId string) bool {
	containers, err := dc.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error(err)
		return false
	}
	for _, container := range containers {
		if container.ID == containerId {
			return true
		}
	}
	return false
}

func getmgctlBinFile() (*os.File, error) {
	mgctlFile := viper.GetString("mgctlTar")
	if _, err := os.Stat(mgctlFile); err == nil {
		return os.Open(mgctlFile)
	} else {
		return nil, err
	}
}

func copymgctlToContainerdContainer(containerId string) {
	c, err := containerd.New("/var/run/containerd/containerd.sock")
	defer c.Close()
	if err != nil {
		log.Error(err)
	}
	if !isContainerdContainerExists(c, containerId) {
		log.Warnf("conatiner %s does not exists", containerId)
		return
	}
	mntLocation, err := mountContainerdContainerFS(c, containerId)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err := os.Stat(mntLocation + "/usr/bin/mgctl"); err == nil {
		return
	}

	log.Warnf("mgctl not found")
	f, err := getmgctlBinFile()
	if err != nil {
		log.Error(err)
		return
	}
	defer f.Close()
	dst, err := os.Create(mntLocation + "/usr/bin/mgctl")
	if err != nil {
		log.Error(err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, f); err != nil {
		log.Error(err)
	}

	unmountContainerdContainerFS(mntLocation)

}

func isContainerdContainerExists(c *containerd.Client, containerId string) bool {
	ctx := namespaces.WithNamespace(context.Background(), "default")
	in := &tasks.ListTasksRequest{}
	tc := c.TaskService()
	r, err := tc.List(ctx, in)
	if err != nil {
		log.Error(err)
		return false
	}
	for _, t := range r.Tasks {
		if t.ContainerID == containerId {
			return true
		}
	}
	return false
}

func mountContainerdContainerFS(c *containerd.Client, containerId string) (mntLocation string, err error) {
	sc := c.SnapshotService(containerd.DefaultSnapshotter)
	ctx := namespaces.WithNamespace(context.Background(), "default")
	mounts, err := sc.Mounts(ctx, containerId)
	if err != nil {
		return "", err
	}
	mountCmd := ""
	tmpMountLocation := fmt.Sprintf("/tmp/%s", containerId)
	for _, m := range mounts {
		mountCmd = fmt.Sprintf("mount -t overlay overlay /tmp/%s -o %s", tmpMountLocation, strings.Join(m.Options, ","))
	}

	if mountCmd == "" {
		return "", errors.New("error mounting container FS to host, mountCmd is empty")
	}

	if err := exec.Command(mountCmd).Run(); err != nil {
		return "", err
	}

	return tmpMountLocation, nil
}

func unmountContainerdContainerFS(mntLocation string) {
	if err := exec.Command(fmt.Sprintf("umount %s", mntLocation)).Run(); err != nil {
		log.Error(err)
	}
}
