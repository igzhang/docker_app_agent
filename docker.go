package main

import (
	"context"
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func operateCMD(cmd *ReceiveCMD) error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	switch cmd.Operate {
	case "compose":
		if err := pullImage(&ctx, cli, cmd.Image, cmd.RegistryAuth); err != nil {
			log.Printf("pull container failed: %v\n", err)
			return err
		}
		containerID, err := getContainerID(&ctx, cli, cmd.AppName)
		if err != nil {
			log.Printf("get container failed: %v\n", err)
			return err
		}
		if len(containerID) > 0 {
			if err = removeContainer(&ctx, cli, containerID); err != nil {
				log.Printf("remove container failed: %v\n", err)
				return err
			}
		}
		_, err = createAndStartContainer(&ctx, cli, cmd.AppName, cmd.Image, &cmd.Extra)
		if err != nil {
			log.Printf("start container failed: %v\n", err)
			return err
		}
	case "stop":
		containerID, err := getContainerID(&ctx, cli, cmd.AppName)
		if err != nil {
			log.Printf("get container failed: %v\n", err)
			return err
		}
		if len(containerID) > 0 {
			if err = removeContainer(&ctx, cli, containerID); err != nil {
				log.Printf("remove container failed: %v\n", err)
				return err
			}
		}
	}
	defer cli.Close()
	return nil
}

func pullImage(ctx *context.Context, client *client.Client, imageName string, registryAuth string) error {
	pullOptions := image.PullOptions{}
	if registryAuth != "" {
		pullOptions.RegistryAuth = registryAuth
	}
	out, err := client.ImagePull(*ctx, imageName, pullOptions)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
	return nil
}

func getContainerID(ctx *context.Context, client *client.Client, containerName string) (string, error) {
	containers, err := client.ContainerList(*ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, c := range containers {
		if c.Names[0][1:] == containerName {
			return c.ID, nil
		}
	}
	return "", nil
}

func removeContainer(ctx *context.Context, client *client.Client, containerID string) error {
	if err := client.ContainerRemove(*ctx, containerID, container.RemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func createAndStartContainer(ctx *context.Context, client *client.Client, containerName string, imageName string, configDetail *ContainerConfigDetails) (string, error) {
	containerConfig := &container.Config{
		Image: imageName,
	}
	if len(configDetail.Cmd) > 0 {
		containerConfig.Cmd = configDetail.Cmd
	}
	if len(configDetail.Env) > 0 {
		containerConfig.Env = configDetail.Env
	}

	hostConfig := &container.HostConfig{}
	if len(configDetail.Port) > 0 {
		portBind := nat.PortMap{}
		for _, portStr := range configDetail.Port {
			portBindSlice := strings.SplitN(portStr, ":", 2)
			portBind[nat.Port(portBindSlice[1])] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: portBindSlice[0]}}
		}
		hostConfig.PortBindings = portBind
	}
	if len(configDetail.Mount) > 0 {
		mounts := []mount.Mount{}
		for _, mountStr := range configDetail.Mount {
			mountSlice := strings.SplitN(mountStr, ":", 2)
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: mountSlice[0],
				Target: mountSlice[1],
			})
		}
		hostConfig.Mounts = mounts
	}

	if configDetail.Cpu > 0 {
		hostConfig.Resources.NanoCPUs = int64(configDetail.Cpu * 1e9)
	}
	if configDetail.Memory > 0 {
		hostConfig.Resources.Memory = int64(configDetail.Memory * 1024 * 1024)
	}

	resp, err := client.ContainerCreate(*ctx, containerConfig, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", err
	}

	if err := client.ContainerStart(*ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}
