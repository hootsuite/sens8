package check

type ExitStatus uint8

const (
	Success ExitStatus = iota
	Warning
	Error
)

type CheckOutput struct {
	*CheckRequest
	Status   ExitStatus `json:"status"`
	Output   string     `json:"output"`
	Duration float64    `json:"duration,omitempty"`
	Executed int64      `json:"executed,omitempty"`
}
