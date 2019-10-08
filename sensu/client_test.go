package sensu

import (
	"encoding/json"
	"testing"

	"github.com/hootsuite/sens8/internal/checks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	defaults = map[string]interface{}{
		"occurrences": float64(3),
		"handlers":    []interface{}{"default"},
		"pager_team":  "production_engineering",
	}
)

type MockTransport struct {
	connected    bool
	exchangeType string
	exchangeName string
	key          string
	message      []byte
}

func (m *MockTransport) Connect() error {
	m.connected = true
	return nil
}
func (m *MockTransport) IsConnected() bool {
	return m.connected
}
func (m *MockTransport) Close() error {
	m.connected = false
	return nil
}
func (m *MockTransport) Publish(exchangeType, exchangeName, key string, message []byte) error {
	m.exchangeType = exchangeType
	m.exchangeName = exchangeName
	m.key = key
	m.message = message
	return nil
}
func (m *MockTransport) Subscribe(key, exchangeName, queueName string, messageChan chan []byte, stopChan chan bool) error {
	return nil
}
func (m *MockTransport) GetClosingChan() chan bool {
	return make(chan bool)
}

func TestNewSensuClient(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)

	assert.Equal("sens8", s.Config.ClientConfig.Name)
	assert.Equal([]string{}, s.Config.ClientConfig.Subscriptions)
	assert.Equal("deregister_client", s.Config.ClientConfig.Deregistration.Handler)
	assert.False(s.Transport.IsConnected())
}

func TestNewSensuClientExtras(t *testing.T) {
	t.Skip()
	//@todo - check unexported fields
}

func TestNewSensuClientDefaults(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)
	assert.EqualValues(defaults, s.Config.Defaults)
}

func TestNewSensuClientKeepalive(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)

	assert.Equal(&[]string{"default"}, s.Config.ClientConfig.Keepalive.Handlers)
	assert.Equal("", s.Config.ClientConfig.Keepalive.Handler)
	assert.Equal(40, s.Config.ClientConfig.Keepalive.Thresholds.Warning)
	assert.Equal(60, s.Config.ClientConfig.Keepalive.Thresholds.Critical)
}

func TestSensuClient_Start(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)
	tr := MockTransport{}
	tr.Connect()
	s.Transport = &tr

	s.postKeepAlive()
	assert.True(len(tr.message) > 0)

	msg := make(map[string]interface{})
	err = json.Unmarshal(tr.message, &msg)
	require.NoError(err)

	require.Contains(msg, "name")
	assert.Equal("sens8", msg["name"])

	require.Contains(msg, "pager_team")
	assert.Equal("production_engineering", msg["pager_team"])

	require.Contains(msg, "thresholds")
	require.IsType(map[string]interface{}{}, msg["thresholds"])
	thresholds := map[string]interface{}{"warning": float64(40), "critical": float64(60)}
	assert.EqualValues(thresholds, msg["thresholds"].(map[string]interface{}))
}

func TestSensuClient_Deregister(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)
	tr := MockTransport{}
	tr.Connect()
	s.Transport = &tr

	s.Deregister("test_client")
	require.NoError(err)
	assert.True(len(tr.message) > 0)

	msg := make(map[string]interface{})
	err = json.Unmarshal(tr.message, &msg)
	require.NoError(err)

	require.Contains(msg, "check")
	require.IsType(map[string]interface{}{}, msg["check"])
	ch := msg["check"].(map[string]interface{})

	require.Contains(ch, "source")
	assert.Equal("test_client", ch["source"])

	require.Contains(ch, "handler")
	assert.Equal("deregister_client", ch["handler"])

	require.Contains(ch, "name")
	assert.Equal("deregistration", ch["name"])
}

func TestSensuClient_PostCheckResult(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	s, err := New("testdata/conf.json")
	require.NoError(err)
	tr := MockTransport{}
	tr.Connect()
	s.Transport = &tr

	source := "test_client"
	res := checks.NewCheckResultFromConfig(checks.CheckConfig{
		Name:     "test_check",
		Command:  "foo",
		Interval: 1,
		Source:   &source,
	})
	res.Output = "test_output"
	res.Status = checks.WARN

	err = s.PostCheckResult(res)
	require.NoError(err)

	msg := make(map[string]interface{})
	err = json.Unmarshal(tr.message, &msg)
	require.NoError(err)

	require.Contains(msg, "check")
	require.IsType(map[string]interface{}{}, msg["check"])
	ch := msg["check"].(map[string]interface{})

	require.Contains(ch, "source")
	assert.Equal(source, ch["source"])

	require.Contains(ch, "output")
	assert.Equal("test_output", ch["output"])

	require.Contains(ch, "status")
	assert.Equal(float64(checks.WARN), ch["status"])
}
