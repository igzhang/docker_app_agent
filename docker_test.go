package main

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
)

func TestPullImage(t *testing.T) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Error(err)
	}
	if err := pullImage(&ctx, cli, "busybox", ""); err != nil {
		t.Errorf("pull container failed: %v\n", err)
	}
}

func TestGetContainerIDNotFound(t *testing.T) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Error(err)
	}
	containerID, err := getContainerID(&ctx, cli, "b")
	if err != nil {
		t.Errorf("get container id failed: %v\n", err)
	}
	if len(containerID) != 0 {
		t.Error("get container id not expected")
	}
}

func TestCreateAndRemoveContainer(t *testing.T) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Error(err)
	}
	detail := &ContainerConfigDetails{Cmd: []string{"sleep", "20"}}
	containerID, err := createAndStartContainer(&ctx, cli, "aaa", "busybox", detail)
	if err != nil {
		t.Error(err)
	}

	foundContainerID, err := getContainerID(&ctx, cli, "aaa")
	if err != nil {
		t.Errorf("get container id failed: %v\n", err)
	}
	if containerID != foundContainerID {
		t.Error("get container id not expected")
	}

	if err := removeContainer(&ctx, cli, foundContainerID); err != nil {
		t.Errorf("delete container failed: %v\n", err)
	}

	foundContainerID, err = getContainerID(&ctx, cli, "aaa")
	if err != nil {
		t.Errorf("get container id failed: %v\n", err)
	}
	if len(foundContainerID) != 0 {
		t.Error("get container id not expected")
	}

}
