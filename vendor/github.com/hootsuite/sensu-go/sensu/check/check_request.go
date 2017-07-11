package check

type CheckRequest struct {
	*Check
	Issued int64 `json:"issued,omitempty"`
}
