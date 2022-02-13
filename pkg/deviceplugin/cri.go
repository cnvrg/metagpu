package deviceplugin

import (
	"context"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
)

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
		_ = f.Close()
	}
}

func shouldCopy(dc *docker.Client, containerId string) bool {
	_, err := dc.ContainerStatPath(context.Background(), containerId, "/usr/bin/mgctl")
	if err != nil {
		log.Warnf("mgctl not found, err: %s", err)
		return true
	}
	return false
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

func getmgctlBinFile() *os.File {
	mgctlFile := viper.GetString("mgctlTar")
	if _, err := os.Stat(mgctlFile); err == nil {
		if f, err := os.Open(mgctlFile); err == nil {
			return f
		} else {
			log.Error(err)
		}
	} else {
		log.Error(err)
	}

	return nil
}
