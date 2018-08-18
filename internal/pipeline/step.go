package pipeline

const (
	STEP_WAITING = iota
	STEP_RUNNING
	STEP_FAILURE
	STEP_SUCCESS
)

type Step struct {
	Name       string
	Type       string
	Entrypoint string
	Command    string
	Directory  string
	User       string
	Timeout    int
	Env        map[string]string
}
