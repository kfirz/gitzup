package util

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var cli *client.Client = nil

func init() {
	var err error
	cli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}

type DockerRequest struct {
	Image         string
	ContainerName string
	Env           []string
	Volumes       map[string]struct{}
	Input         interface{}
}

type DockerError struct {
	error
	Msg     string
	Cause   error
	Request DockerRequest
}

func (e *DockerError) Error() string {
	return e.Msg + ": " + e.Cause.Error() + fmt.Sprintf(" (%s)", e.Request.ContainerName)
}

func (request *DockerRequest) Invoke(schema *Schema, response interface{}) error {
	ctx := context.Background()
	defer ctx.Done()

	// check if we already have this image
	imageListArgs := filters.NewArgs()
	imageListArgs.Add("reference", request.Image)
	imageList, err := cli.ImageList(ctx, types.ImageListOptions{Filters: imageListArgs})
	if err != nil {
		return &DockerError{Msg: "could not fetch local image list", Cause: err, Request: *request}
	} else if len(imageList) == 0 {
		log.Printf("Pulling image '%s'...\n", request.Image)
		reader, err := cli.ImagePull(ctx, request.Image, types.ImagePullOptions{All: true})
		if err != nil {
			return &DockerError{Msg: "could not pull image", Cause: err, Request: *request}
		}
		defer reader.Close()
		// TODO: pretty-print image-pull reader messages to stderr
		if io.Copy(os.Stderr, reader); err != nil {
			return &DockerError{Msg: "could not stream image pull progress", Cause: err, Request: *request}
		}
	}

	// create container
	c, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Domainname:   "gitzup.local",
			AttachStdin:  false,
			AttachStdout: true,
			AttachStderr: true,
			OpenStdin:    true,
			StdinOnce:    true,
			Env:          request.Env,
			Image:        request.Image,
			Volumes:      request.Volumes,
		},
		&container.HostConfig{AutoRemove: false},
		nil,
		request.ContainerName)
	if err != nil {
		return &DockerError{Msg: "could not create container", Cause: err, Request: *request}
	}
	defer func() {
		if err := cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{}); err != nil {
			log.Printf("Failed removing container '%s': %v", c.ID, err)
		}
	}()

	// start container
	if err := cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{}); err != nil {
		return &DockerError{Msg: "could not start container", Cause: err, Request: *request}
	}

	// if input provided, attach to container and send the input to stdin
	if request.Input != nil {
		resp, err := cli.ContainerAttach(ctx, c.ID, types.ContainerAttachOptions{Stream: true, Stdin: true})
		if err != nil {
			return &DockerError{Msg: "could not attach to container", Cause: err, Request: *request}
		}
		defer resp.Close()

		// send input to the container's stdin
		bytes, err := json.Marshal(request.Input)
		if err != nil {
			return &DockerError{Msg: "could not send input to container", Cause: err, Request: *request}
		}
		resp.Conn.Write(bytes)
		resp.CloseWrite()
	}

	// stream logs to our stdout
	out, err := cli.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})
	defer out.Close()
	if err != nil {
		return &DockerError{Msg: "could not stream back container logs", Cause: err, Request: *request}
	}
	io.Copy(os.Stdout, out)

	// wait for container to finish
	ctx30Sec, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if _, err := cli.ContainerWait(ctx30Sec, c.ID); err != nil {
		return &DockerError{Msg: "could not wait for container to exit", Cause: err, Request: *request}
	}

	// read JSON response from resource (fail if missing or invalid)
	reader, _, err := cli.CopyFromContainer(ctx, c.ID, "/gitzup/output.json")
	if err != nil {
		return &DockerError{Msg: "could not stat path '/gitzup/output.json' path inside container", Cause: err, Request: *request}
	}
	defer reader.Close()
	tr := tar.NewReader(reader)
	_, err = tr.Next()
	if err == io.EOF {
		return &DockerError{Msg: "protocol error: could not find '/gitzup/output.json' inside container", Cause: err, Request: *request}
	} else if err != nil {
		return &DockerError{Msg: "could not read path '/gitzup/output.json' path inside container", Cause: err, Request: *request}
	}
	bytes, err := ioutil.ReadAll(tr)
	if err != nil {
		return &DockerError{Msg: "could not read path '/gitzup/output.json' path inside container", Cause: err, Request: *request}
	}
	err = schema.ParseAndValidate(response, bytes)
	if err != nil {
		return &DockerError{Msg: "failed parsing or validating container response", Cause: err, Request: *request}
	}
	return nil
}
