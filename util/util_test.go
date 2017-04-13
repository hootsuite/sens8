package util_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/hootsuite/sens8/util"

	"time"
)

func TestDecodeWithExtraFields(t *testing.T) {
	assert := assert.New(t)

	type T struct {
		Defined string
		Default int
	}
	data := map[string]interface{}{"Defined": "overridden", "Extra": "extra"}
	obj := T{Defined: "override", Default: 1}

	extras, err := util.DecodeWithExtraFields(data, &obj)

	assert.NoError(err)
	assert.Len(extras, 1)
	assert.Equal("extra", extras["Extra"], "wrong data in extras")
	assert.NotEqual("override", obj.Defined, "new data should overwrite defaults")
	assert.Equal(1, obj.Default, "defaults should persist if not overwritten")
}

func TestJsonStructToMap(t *testing.T) {
	assert := assert.New(t)

	s := struct{
		Defined string `json:"defined"`
		Default string `json:"default"`
		Blank string `json:"-"`
	}{
		Defined: "overridden",
		Default: "defaultval",
	}
	m := map[string]interface{}{"defined": "override"}
	err := util.JsonStructToMap(s, &m)

	assert.NoError(err)
	assert.Equal("overridden", m["defined"], "new data should overwrite defaults")
	assert.Equal("defaultval", m["default"], "defaults should persist if not overwritten")
	assert.NotContains(m, "blank")
}

func TestSecondsSince(t *testing.T) {
	n := time.Now()
	time.Sleep(time.Millisecond)
	s := util.SecondsSince(n)
	assert.InDelta(t, s, 0.001, 0.001)
}
