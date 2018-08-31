package buildagent

import (
	"fmt"
	"github.com/kfirz/gitzup/internal/protocol"
	"github.com/kfirz/gitzup/internal/util"
	"log"
)

var protocolInitResponseSchema *util.Schema

func init() {
	schema, err := util.NewSchema("protocol/init.response.schema.json")
	if err != nil {
		panic(err)
	}
	protocolInitResponseSchema = schema
}

type InitResponse struct {
}

func (resource *Resource) Initialize() error {
	log.Printf("Initializing resource '%s'...\n", resource.Name)
	initRequest := util.DockerRequest{
		Input: protocol.InitRequest{
			RequestId: resource.Request.Id,
			Resource: protocol.InitRequestResourceInfo{
				Type: resource.Type,
				Name: resource.Name,
			},
		},
		ContainerName: "init-" + resource.Name,
		Env: []string{
			"GITZUP=true",
			fmt.Sprintf("RESOURCE_NAME=%s", resource.Name),
			fmt.Sprintf("RESOURCE_TYPE=%s", resource.Type),
			// TODO: send Gitzup version in resource init env
		},
		Volumes: map[string]struct{}{
			fmt.Sprintf("%s/cache:/tmp/gitzup/cache:rw", resource.Workspace()): {},
		},
		Image: resource.Type,
	}

	var initResponse InitResponse
	if err := initRequest.Invoke(protocolInitResponseSchema, initResponse); err != nil {
		return err
	}

	log.Printf("Init response: %+v", initResponse)

	return nil
}
