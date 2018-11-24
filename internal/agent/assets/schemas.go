package assets

import (
	commonAssets "github.com/kfirz/gitzup/internal/common/assets"
)

var resourceSchema = commonAssets.LoadSchema(localAsset("internal/agent/assets/resource.json"))
var buildRequestSchema = commonAssets.LoadSchema(localAsset("internal/agent/assets/build.request.json"), localAsset("internal/agent/assets/resource.json"))
var buildResponseSchema = commonAssets.LoadSchema(localAsset("internal/agent/assets/build.response.json"))

func GetResourceActionSchema() *commonAssets.Schema       { return commonAssets.GetResourceActionSchema() }
func GetResourceSchema() *commonAssets.Schema             { return resourceSchema }
func GetBuildRequestSchema() *commonAssets.Schema         { return buildRequestSchema }
func GetBuildResponseSchema() *commonAssets.Schema        { return buildResponseSchema }
func GetResourceInitRequestSchema() *commonAssets.Schema  { return commonAssets.GetResourceInitRequestSchema() }
func GetResourceInitResponseSchema() *commonAssets.Schema { return commonAssets.GetResourceInitResponseSchema() }

func localAsset(source string) []byte {
	schemaBytes, err := Asset(source)
	if err != nil {
		panic(err)
	}
	return schemaBytes
}
