package check

type CheckType string

const (
	Metric   CheckType = "metric"
	Standard CheckType = "standard"
)

type Check struct {
	Name         string    `json:"name,omitempty"`
	Type         CheckType `json:"type,omitempty"`
	Command      string    `json:"command,omitempty"`
	Extension    string    `json:"extension,omitempty"`
	Standalone   bool      `json:"standalone,omitempty"`
	Subscribers  []string  `json:"subscribers,omitempty"`
	Handler      string    `json:"handler,omitempty"`
	Handlers     []string  `json:"handlers,omitempty"`
	Source       string    `json:"source,omitempty"`
	Interval     int64     `json:"interval,omitempty"`
	Occurrences  int64     `json:"occurrences,omitempty"`
	Refresh      int64     `json:"refresh,omitempty"`
	Dependencies []string  `json:"dependencies,omitempty"`
	Notification string    `json:"notification,omitempty"`
}
