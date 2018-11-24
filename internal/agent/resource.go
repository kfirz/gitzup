package agent

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/kfirz/gitzup/internal/agent/assets"
	"github.com/kfirz/gitzup/internal/common"
	commonAssets "github.com/kfirz/gitzup/internal/common/assets"
)

// Represents a single resource in a build request.
type Resource interface {
	Request() Request
	Name() string
	Type() string
	Config() interface{}
	ConfigSchema() *commonAssets.Schema
	WorkspacePath() string
	Init(ctx context.Context) error
	DiscoverState(ctx context.Context) error
	Apply(ctx context.Context) error
}

type resourceImpl struct {
	request         Request
	name            string
	resourceType    string
	resourceConfig  interface{}
	workspacePath   string
	configSchema    *commonAssets.Schema
	initAction      Action
	discoveryAction Action
}

type resourceInitRequest struct {
	Id       string   `json:"id"`
	Resource Resource `json:"resource"`
}

func (res *resourceImpl) Request() Request {
	return res.request
}

func (res *resourceImpl) Name() string {
	return res.name
}

func (res *resourceImpl) Type() string {
	return res.resourceType
}

func (res *resourceImpl) Config() interface{} {
	return res.resourceConfig
}

func (res *resourceImpl) ConfigSchema() *commonAssets.Schema {
	return res.configSchema
}

func (res *resourceImpl) WorkspacePath() string {
	return res.workspacePath
}

func (res *resourceImpl) Init(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Initializing resource")

	// initialize the resource
	var response common.ResourceInitResponse
	err := res.initAction.Invoke(
		ctx,
		&resourceInitRequest{
			Id:       res.Request().Id(),
			Resource: res,
		},
		assets.GetResourceInitResponseSchema(),
		&response,
	)
	if err != nil {
		return errors.WrapPrefix(err, "failed initializing resource", 0)
	}

	// build the resource configuration schema
	resourceConfigSchema, err := commonAssets.NewSchema(
		response.ConfigSchema,
		assets.GetResourceActionSchema(),
		assets.GetResourceSchema(),
		assets.GetBuildRequestSchema(),
		assets.GetBuildResponseSchema(),
		assets.GetResourceInitRequestSchema(),
		assets.GetResourceInitResponseSchema(),
	)
	if err != nil {
		return err
	}
	res.configSchema = resourceConfigSchema

	// use the configuration schema to validate the resource's configuration
	if err := res.configSchema.Validate(res.Config()); err != nil {
		return err
	}

	// read and set the resource's state discovery action
	res.discoveryAction = &actionImpl{
		resource:   res,
		name:       "state",
		image:      response.StateAction.Image,
		entrypoint: response.StateAction.Entrypoint,
		cmd:        response.StateAction.Cmd,
	}

	return nil
}

func (res *resourceImpl) DiscoverState(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Discovering state")

	// TODO eric: implement
	panic("not implemented")
}

func (res *resourceImpl) Apply(ctx context.Context) error {
	ctx = context.WithValue(ctx, "resource", res.Name())

	From(ctx).Info("Applying")

	// TODO eric: implement
	panic("not implemented")
}
