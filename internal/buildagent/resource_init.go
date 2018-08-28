package buildagent

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
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
		fmt.Sprintf("%s/cache:/tmp/gitzup/cache:rw", resource.Workspace()): {},
	}
}

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...\n", resource.Name)
	ctx := context.Background()
	defer ctx.Done()

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	log.Printf("Pulling Docker image '%s'...", resource.Type)
	reader, err := dockerClient.ImagePull(ctx, resource.Type, types.ImagePullOptions{
		All: true,
	})
	defer reader.Close()
	io.Copy(ioutil.Discard, reader)
	if err != nil {
		return err
	}

	// create container
	containerName := "init-" + resource.Name
	log.Printf("Creating container '%s'...", containerName)
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
		containerName)
	if err != nil {
		return err
	}
	defer func() {
		log.Printf("Removing container '%s'", c.ID)
		err = dockerClient.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
		if err != nil {
			log.Printf("Failed removing container '%s': %#v", c.ID, err)
		}
	}()

	// start container
	log.Printf("Starting container '%s'...", c.ID)
	if err := dockerClient.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// stream logs to our stdout
	log.Printf("Fetching logs for container '%s'...", c.ID)
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
	log.Printf("Waiting for container '%s' to exit... (timeout in 30 seconds)", c.ID)
	if _, err := dockerClient.ContainerWait(ctx30Sec, c.ID); err != nil {
		return err
	}

	return nil
}
