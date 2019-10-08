package checks_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/hootsuite/sens8/internal/checks"
)

var (
	minimal = `[{
			"name": "minimal",
			"command": "test_check",
			"interval": 1
		}]`
	multiple = `[{
			"name": "minimal",
			"command": "test_check",
			"interval": 1
		},{
			"name": "minimal2",
			"command": "test_check",
			"interval": 1
		}]`
	partialValid = `[{
			"name": "minimal",
			"command": "test_check",
			"interval": 1
		},{
			"derp": "100%"
		}]`
	arbitraryFields = `[{
			"name": "minimal",
			"command": "test_check",
			"interval": 1,
			"foo": "bar"
		}]`
	argv = `[{
			"name": "minimal",
			"command": "test_check -a 1 --bar \"2 3\"",
			"interval": 1
		}]`
	garbage      = `["adsf}]`
	noName       = `[{"command": "test_check", "interval": 1}]`
	noCommand    = `[{"name": "minimal", "interval": 1}]`
	noInterval   = `[{"name": "minimal", "command": "test_check"}]`
	zeroInterval = `[{"name": "minimal", "command": "test_check", "interval": 0}]`
)

type TestCheck struct {
	checks.BaseCheck
}

func NewTestCheck(config checks.CheckConfig) (checks.Check, error) {
	c := TestCheck{}
	c.Config = config
	return &c, nil
}

func (c *TestCheck) Usage() checks.CheckUsage {
	return checks.CheckUsage{
		Description: "description",
		Flags:       "flags",
	}
}

func (c *TestCheck) Update(resource interface{}) {}

func (c *TestCheck) Execute() (checks.CheckResult, error) {
	return checks.CheckResult{}, nil
}
func init() {
	checks.RegisterCheck("test_check", NewTestCheck, []string{"deployment"})
}

func TestNewCheck(t *testing.T) {
	assert := assert.New(t)
	conf := checks.CheckConfig{Name: "test_name", Command: "test_check", Interval: 1, Id: "test_check"}
	c, err := checks.NewCheck(conf, "deployment")
	assert.NoError(err)
	assert.IsType(&TestCheck{}, c)
	assert.Equal("test_name", c.GetConfig().Name)
}

func TestNewCheckFiltered(t *testing.T) {
	assert := assert.New(t)
	conf := checks.CheckConfig{Name: "test_name", Command: "test_check", Interval: 1, Id: "test_check"}
	_, err := checks.NewCheck(conf, "pod")
	assert.Error(err)
}

func TestRegisterCheckTwice(t *testing.T) {
	assert := assert.New(t)
	f := func(config checks.CheckConfig) (checks.Check, error) {
		return &TestCheck{}, nil
	}
	err := checks.RegisterCheck("reg_twice", f, []string{"deployment"})
	assert.NoError(err)
	err = checks.RegisterCheck("reg_twice", f, []string{"deployment"})
	assert.Error(err)
}

func TestParseCheckConfigsMinimal(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	checks, err := checks.ParseCheckConfigs(minimal, "testSource", "deployment")

	assert.Len(err, 0)
	require.Len(checks, 1)
	c := checks[0]
	assert.IsType(&TestCheck{}, c)
	assert.NotNil(c.GetConfig())
	assert.Equal("minimal", c.GetConfig().Name)
	assert.Equal("test_check", c.GetConfig().Command)
	assert.Equal(1, c.GetConfig().Interval)
	assert.Equal("testSource", *c.GetConfig().Source)
	assert.Equal("test_check", c.GetConfig().Id, "has valid hash")
	require.Len(c.GetConfig().Argv, 1)
	assert.Equal("test_check", c.GetConfig().Argv[0])
	assert.True(c.GetHash() > 0, "has valid hash")
}

func TestParseCheckConfigsWithDefaults(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	checks.Defaults = map[string]interface{}{"interval": 123, "foo": "bar"}
	defer func() {
		checks.Defaults = make(map[string]interface{})
	}()

	checks, err := checks.ParseCheckConfigs(minimal, "testSource", "deployment")

	assert.Len(err, 0)
	require.Len(checks, 1)
	c := checks[0]
	assert.Equal(1, c.GetConfig().Interval, "check config should overwrite defaults")
	if assert.Contains(c.GetConfig().ExtraFields, "foo") {
		assert.Equal("bar", c.GetConfig().ExtraFields["foo"])
	}
}

func TestParseCheckConfigsValidFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for _, i := range []string{noName, noCommand, noInterval, zeroInterval} {
		_, err := checks.ParseCheckConfigs(i, "testSource", "deployment")
		require.Len(err, 1)
		assert.Error(err[0])
	}
}

func TestParseCheckConfigsPartialValid(t *testing.T) {
	assert := assert.New(t)

	checks, err := checks.ParseCheckConfigs(partialValid, "testSource", "deployment")
	assert.Len(err, 1)
	assert.Len(checks, 1)
}

func TestParseCheckConfigsWithArbitraryFields(t *testing.T) {
	assert := assert.New(t)

	checks, err := checks.ParseCheckConfigs(arbitraryFields, "testSource", "deployment")
	assert.Len(err, 0)
	if assert.Len(checks, 1) {
		c := checks[0]
		if assert.Contains(c.GetConfig().ExtraFields, "foo") {
			assert.Equal("bar", c.GetConfig().ExtraFields["foo"])
		}
	}
}

func TestParseCheckConfigsWithMultiple(t *testing.T) {
	assert := assert.New(t)
	checks, err := checks.ParseCheckConfigs(multiple, "testSource", "deployment")
	assert.Len(err, 0)
	assert.Len(checks, 2)
}

func TestParseCheckConfigsWithGarbage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	_, err := checks.ParseCheckConfigs(garbage, "testSource", "deployment")
	require.Len(err, 1)
	assert.Error(err[0])
}

func TestParseCheckConfigsArgv(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	checks, err := checks.ParseCheckConfigs(argv, "testSource", "deployment")

	assert.Len(err, 0)
	require.Len(checks, 1)
	c := checks[0]
	assert.Len(c.GetConfig().Argv, 5)
	assert.EqualValues([]string{"test_check", "-a", "1", "--bar", "2 3"}, c.GetConfig().Argv)
}

func TestNewCheckResultFromConfig(t *testing.T) {
	assert := assert.New(t)
	conf := checks.CheckConfig{Name: "test_name", Command: "test_check", Interval: 1, Id: "test_check"}
	res := checks.NewCheckResultFromConfig(conf)
	assert.Equal("test_name", res.Name)
	assert.Equal("test_check", res.Command)
	assert.Equal("test_check", res.Id)
	assert.Equal(1, res.Interval)
	assert.True(res.Executed <= time.Now().Unix())
	assert.True(res.Issued <= time.Now().Unix())
}

func TestCheckResult_JsonResponse(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	conf := checks.CheckConfig{Name: "test_name", Command: "test_check", Interval: 1, Id: "test_check"}
	res := checks.NewCheckResultFromConfig(conf)
	buf, err := res.JsonResponse("client_name")
	require.NoError(err)

	s := make(map[string]interface{})
	err = json.Unmarshal(buf, &s)
	assert.NoError(err)

	require.Contains(s, "client")
	assert.Equal("client_name", s["client"])
	require.Contains(s, "check")
	require.Contains(s["check"], "name")
	require.IsType(map[string]interface{}{}, s["check"])
	assert.Equal(s["check"].(map[string]interface{})["name"], "test_name")
}

func TestDocsComplete(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	d := checks.Docs()
	require.True(len(d) > 0, "docs should be non empty")
	ids := checks.CheckFactoryIds()
	for _, id := range ids {
		exists := false
		for _, doc := range d {
			if doc.Id == id {
				exists = true
			}
		}
		assert.True(exists, fmt.Sprintf("doc missing for %s", id))
	}
}

func TestDocsValid(t *testing.T) {
	assert := assert.New(t)
	d := checks.Docs()
	for _, doc := range d {
		assert.NotEmpty(doc.Id, "doc has empty resources")
		assert.NotEmpty(doc.Resources, "doc has empty resources")
		assert.NotEmpty(doc.Flags, "doc has empty flags")
		assert.NotEmpty(doc.Description, "doc has empty description")
	}
}

func TestGenCheckDocs(t *testing.T) {
	assert := assert.New(t)
	d := checks.GenCheckDocs("xxx")
	assert.Exactly("xxx", d)
}

func TestGenCheckDocsMarkdown(t *testing.T) {
	assert := assert.New(t)
	d := checks.GenCheckDocsMarkdown()
	assert.Contains(d, "deployment_status")
}

func TestGenCheckDocsText(t *testing.T) {
	assert := assert.New(t)
	d := checks.GenCheckDocsText()
	assert.Contains(d, "deployment_status")
}
