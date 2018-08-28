package protocol

type InitRequest struct {
	RequestId string                  `json:"requestId"`
	Resource  InitRequestResourceInfo `json:"resource"`
}

type InitRequestResourceInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
