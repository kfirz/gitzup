package buildagent

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kfirz/gitzup/internal/protocol"
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

	// check if we already have this image
	imageListArgs := filters.NewArgs()
	imageListArgs.Add("reference", resource.Type)
	// {"image.name":{"ubuntu":true},"label":{"label1=1":true,"label2=2":true}}
	imageList, err := dockerClient.ImageList(ctx, types.ImageListOptions{Filters: imageListArgs})
	if err != nil {
		return err
	}
	if len(imageList) == 0 {
		log.Printf("Pulling image '%s'...\n", resource.Type)
		reader, err := dockerClient.ImagePull(ctx, resource.Type, types.ImagePullOptions{All: true})
		defer reader.Close()
		if io.Copy(ioutil.Discard, reader); err != nil {
			return err
		}
	} else {
		log.Printf("Image '%s' already present\n", resource.Type)
	}

	// create container
	containerName := "init-" + resource.Name
	c, err := dockerClient.ContainerCreate(
		ctx,
		&container.Config{
			Hostname:     resource.Name,
			Domainname:   "gitzup.local",
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
	if err := dockerClient.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// attach to container
	resp, err := dockerClient.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return err
	}
	defer resp.Close()

	// send init request to the container's stdin
	bytes, err := json.Marshal(protocol.InitRequest{
		RequestId: resource.Request.Id,
		Resource: protocol.InitRequestResourceInfo{
			Type: resource.Type,
			Name: resource.Name,
		},
	})
	if err != nil {
		return err
	}
	resp.Conn.Write(bytes)
	resp.CloseWrite()

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
