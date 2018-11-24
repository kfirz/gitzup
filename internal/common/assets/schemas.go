package assets

var resourceActionSchema = LoadSchema("internal/common/assets/resource.action.json")
var resourceInitRequestSchema = LoadSchema("internal/common/assets/resource.init.request.json")
var resourceInitResponseSchema = LoadSchema("internal/common/assets/resource.init.response.json", "internal/common/assets/resource.action.json")

func GetResourceActionSchema() *Schema       { return resourceActionSchema }
func GetResourceInitRequestSchema() *Schema  { return resourceInitRequestSchema }
func GetResourceInitResponseSchema() *Schema { return resourceInitResponseSchema }
