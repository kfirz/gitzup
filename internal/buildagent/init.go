package buildagent

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/client"
)

func (resource *Resource) Workspace() string {
	return fmt.Sprintf("%s/resources/%s", resource.Request.WorkspacePath(), resource.Name)
}

func (resource *Resource) buildInitEnv() []string {
	return []string{
		"GITZUP=true",
		fmt.Sprintf("RESOURCE_NAME=%s", resource.Name),
		fmt.Sprintf("RESOURCE_TYPE=%s", resource.Type),
		// TODO: send Gitzup version in resource init env
	}
}

func (resource *Resource) buildInitVolumes() map[string]struct{} {
	return map[string]struct{}{
		fmt.Sprintf("%s/cache:/tmp/gitzup/cache:rw", resource.Workspace()): struct{}{},
	}
}

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...", resource)
	ctx := context.Background()

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	reader, err := dockerClient.ImagePull(ctx, resource.Type, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()

	// create container
	c, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Hostname:     resource.Name,
			Domainname:   "gitzup.local",
			Cmd:          []string{"echo", "Hello!"},
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          resource.buildInitEnv(),
			Image:        resource.Type,
			Volumes:      resource.buildInitVolumes(),
		},
		&container.HostConfig{AutoRemove: false},
		nil,
		"init-"+resource.Name)
	if err != nil {
		return err
	}
	defer dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
		Force:         true,
	})

	// start container
	if err := dockerClient.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// stream logs to our stdout
	out, err := dockerClient.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})
	defer out.Close()
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	// wait for container to finish
	ctx30Sec, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err := dockerClient.ContainerWait(ctx30Sec, c.ID); err != nil {
		return err
	}

	return nil
}
