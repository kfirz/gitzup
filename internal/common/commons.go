package common

type Action struct {
	Image      string   `json:"image"`
	Entrypoint []string `json:"entrypoint"`
	Cmd        []string `json:"cmd"`
}

type ResourceInitResponse struct {
	ConfigSchema interface{} `json:"configSchema"`
	StateAction  Action `json:"stateAction"`
}
