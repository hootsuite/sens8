package event

import (
	"encoding/json"
	"strconv"

	"github.com/upfluence/sensu-go/sensu/check"
	"github.com/upfluence/sensu-go/sensu/client"
)

type Client struct {
	Timestamp int64 `json:"timestamp"`
	client.Client
}

type Check struct {
	History    []check.ExitStatus `json:"-"`
	RawHistory []string           `json:"history"`
	Name       string             `json:"name"`
	Command    string             `json:"command"`
	check.CheckOutput
}

type Event struct {
	ID         string `json:"id"`
	Client     Client `json:"client"`
	Check      Check  `json:"check"`
	Occurences uint8  `json:"occurrences"`
	Action     string `json:"action,omitempty"`
}

func UnmarshalEvent(blob []byte) (*Event, error) {
	var (
		event = Event{}
		err   = json.Unmarshal(blob, &event)
	)

	if err == nil {
		for _, rawHistory := range event.Check.RawHistory {
			if i, errConv := strconv.Atoi(rawHistory); errConv == nil {
				event.Check.History = append(event.Check.History, check.ExitStatus(i))
			}
		}
	}

	return &event, err
}
