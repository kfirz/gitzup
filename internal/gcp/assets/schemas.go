package assets

import (
	commonAssets "github.com/kfirz/gitzup/internal/common/assets"
)

var projectConfigSchema = commonAssets.LoadSchema(localAsset("internal/gcp/assets/project.config.json"))

func GetProjectConfigSchema() *commonAssets.Schema       { return projectConfigSchema }

func localAsset(source string) []byte {
	schemaBytes, err := Asset(source)
	if err != nil {
		panic(err)
	}
	return schemaBytes
}
