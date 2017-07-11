package event

import (
	"reflect"
	"testing"

	"github.com/upfluence/sensu-go/sensu/check"
	"github.com/upfluence/sensu-go/sensu/client"
)

type out struct {
	event Event
	err   bool
}

type testCase struct {
	in  string
	out out
}

func TestUnmarshalEvent(t *testing.T) {
	testCases := []testCase{
		testCase{"", out{Event{}, true}},
		testCase{"{}", out{Event{}, false}},
		testCase{
			`
			{
				"id": "foo",
				"client": { "name": "i-424242" },
				"action": "create",
				"check": { "name": "fiz", "history": ["2","2","2"] },
				"occurrences": 27
			}
			`,
			out{
				Event{
					ID:     "foo",
					Action: "create",
					Client: Client{
						0,
						client.Client{Name: "i-424242"},
					},
					Check: Check{
						Name:       "fiz",
						History:    []check.ExitStatus{2, 2, 2},
						RawHistory: []string{"2", "2", "2"},
					},
					Occurences: 27,
				},
				false,
			},
		},
	}

	for _, tCase := range testCases {
		e, err := UnmarshalEvent([]byte(tCase.in))

		if (tCase.out.err && err == nil) || (!tCase.out.err && err != nil) {
			t.Errorf("Wrong error: %s", err.Error())
		}

		if err == nil && !reflect.DeepEqual(*e, tCase.out.event) {
			t.Errorf("Wrong unmarshalling: %+v", *e)
			t.Errorf("Wrong unmarshalling: %+v", tCase.out.event)
		}
	}
}
