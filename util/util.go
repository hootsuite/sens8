package util

import (
	"encoding/json"
	"time"

	"github.com/mitchellh/mapstructure"
)

// DecodeWithExtraFields decodes generic data into
func DecodeWithExtraFields(data map[string]interface{}, v interface{}) (map[string]interface{}, error) {
	var meta mapstructure.Metadata
	extra := make(map[string]interface{})

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata: &meta,
		Result:   v,
	})
	if err != nil {
		return extra, err
	}
	if err = decoder.Decode(data); err != nil {
		return extra, err
	}

	// collect arbitrary fields
	for _, key := range meta.Unused {
		extra[key] = data[key]
	}

	return extra, nil
}

// JsonStructToMap merges s on top of m via `json.Marshal`
// @todo implement this properly with reflection (or maybe not)
func JsonStructToMap(s interface{}, m *map[string]interface{}) error {
	// first turn a to generic json
	generic, err := json.Marshal(s)
	if err != nil {
		return err
	}
	// unmarshal that back to a generic map merging on top of b
	return json.Unmarshal(generic, m)
}

func SecondsSince(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / float64(1000000000)
}

func PadRight(str, pad string, length int) string {
	for {
		if len(str) >= length {
			return str
		}
		str += pad
	}
}
