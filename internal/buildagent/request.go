package buildagent

import "encoding/json"

type Resource struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type BuildRequest struct {
	Resources map[string]Resource `json:"resources"`
}

func NewBuildRequestV1(attributes map[string]string, bytes []byte) (*BuildRequest, error) {
	var req BuildRequest
	err := json.Unmarshal(bytes, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}
