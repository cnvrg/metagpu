package deviceplugin

import (
	"context"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

var mgctlBinLocation = "/usr/bin/mgctl"

func copymgctlToContainer(containerId string) {
	ctx := context.Background()
	cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
	defer cli.Close()
	if err != nil {
		log.Error(err)
		return
	}
	if !isContainerExists(cli, containerId) {
		log.Errorf("container with ID: %s does not exists, can't copy mgctl binary", containerId)
		return
	}

	if !shouldCopy(cli, containerId) {
		return
	}

	if f := getmgctlBinFile(); f != nil {
		if err := cli.CopyToContainer(ctx, containerId, "/usr/bin", f, types.CopyToContainerOptions{}); err != nil {
			log.Error(err)
		}
	}
}

func shouldCopy(dc *docker.Client, containerId string) bool {
	stat, err := dc.ContainerStatPath(context.Background(), containerId, "/usr/bin/mgctl")
	if err != nil {
		log.Error(err)
	}
	log.Info(stat)
	return true
}

func isContainerExists(dc *docker.Client, containerId string) bool {
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

func getmgctlBinFile() io.Reader {
	if _, err := os.Stat(mgctlBinLocation); err == nil {
		if f, err := os.Open(mgctlBinLocation); err == nil {
			return f
		} else {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}

	return nil
}
