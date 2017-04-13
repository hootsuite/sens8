package client

import (
	"time"
	"io/ioutil"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/hootsuite/sensu-go/sensu/transport"
	"github.com/hootsuite/sensu-go/sensu/transport/rabbitmq"
	"github.com/hootsuite/sens8/check"
	"github.com/hootsuite/sens8/util"
)

const (
	connectionTimeout = 5 * time.Second
	keepaliveInterval = 20 * time.Second
	version = "0.28.4" // version is based on what documentation version was referenced
)

type SensuConfig struct {
	RawClient         map[string]interface{}      `json:"client,omitempty"`
	ClientConfig      SensuClientConfig           `json:"-"`
	RabbitMQTransport []*rabbitmq.TransportConfig `json:"rabbitmq,omitempty"`
	Defaults          map[string]interface{}      `json:"defaults,omitempty"`
}

type SensuClientConfig struct {
	SensuClientInfo
	Keepalive KeepAlive `json:"keepalive,omitempty"`
}

type SensuClientInfo struct {
	Name           string              `json:"name"`
	Address        string              `json:"address"`
	Subscriptions  []string            `json:"subscriptions"`
	Deregister     *bool               `json:"deregister"`
	Deregistration *SensuRegistration  `json:"deregistration"`
	Registration   *SensuRegistration  `json:"registration"`
	extraFields	map[string]interface{} `json:"-"`
}

type SensuRegistration struct {
	Handler string `json:"handler"`
}

type KeepAlive struct {
	Handler    string    `json:"handler,omitempty"`
	Handlers   *[]string `json:"handlers,omitempty"`
	Thresholds *struct {
		Warning  int       `json:"warning,omitempty"`
		Critical int       `json:"critical,omitempty"`
	}                    `json:"thresholds,omitempty"`
}

type KeepAliveResponse struct {
	ClientInfo    SensuClientInfo `json:"-"`
	KeepAliveConf KeepAlive       `json:"-"`
	Timestamp     int64           `json:"timestamp"`
	Version       string          `json:"version"`
}

type SensuClient struct {
	Config     SensuConfig
	Transport  transport.Transport
}

// NewSensuClient create a new client based on the confFile
func NewSensuClient(confFile string) (SensuClient, error) {
	s := SensuClient{}

	buf, err := ioutil.ReadFile(confFile)
	if err != nil {
		return s, err
	}

	conf := SensuConfig{}
	if err := json.Unmarshal(buf, &conf); err != nil {
		return s, err
	}

	// decode client
	ci := SensuClientInfo{}
	extras, err := util.DecodeWithExtraFields(conf.RawClient, &ci)
	if err != nil {
		return s, err
	}
	delete(extras, "keepalive")
	ci.extraFields = extras

	// decode keepalive
	ka := KeepAlive{}
	if raw, ok := conf.RawClient["keepalive"]; ok {
		_, err := util.DecodeWithExtraFields(raw.(map[string]interface{}), &ka)
		if err != nil {
			return s, err
		}
	}

	conf.ClientConfig = SensuClientConfig{
		SensuClientInfo: ci,
		Keepalive: ka,
	}

	s.Config = conf
	s.Transport = rabbitmq.NewRabbitMQHATransport(conf.RabbitMQTransport)

	return s, nil
}

// Start starts the client & underlying transport
func (c *SensuClient) Start(stopCh chan struct{}) {
	glog.Info("Starting sensu client")

	for {
		c.Transport.Connect()

		// reconnect
		for !c.Transport.IsConnected() {
			select {
			case <-time.After(connectionTimeout):
				c.Transport.Connect()
			case <-stopCh:
				c.Transport.Close()
				return
			}
		}

		// wait for close
		select {
		case <-stopCh:
			c.Transport.Close()
			return
		case <-c.Transport.GetClosingChan():
			c.Transport.Close()
		}
	}
}

// StartKeepalive starts the keepalive heartbeats for sens8 itself which registers it in sensu
func (c *SensuClient) StartKeepalive(stopCh chan struct{}) {
	t := time.Tick(keepaliveInterval)
	c.postKeepAlive()
	for {
		select {
		case <-t:
			c.postKeepAlive()
		case <-stopCh:
			return
		}
	}
}

// PostCheckResult send out a check result to rabbitmq
// @todo buffer sending? This might be bad as timestamps & keepalives could be affected
func (c *SensuClient) PostCheckResult(result check.CheckResult) error {
	buf, err := result.JsonResponse(c.Config.ClientConfig.Name)
	if err != nil {
		return err
	}
	glog.V(6).Infof("PostCheckResult: %s", string(buf))
	return c.Transport.Publish("direct", "results", "", buf)
}

// postKeepAlive sends out keepalive heartbeats
func (c *SensuClient) postKeepAlive() error {
	res := KeepAliveResponse{
		ClientInfo: c.Config.ClientConfig.SensuClientInfo,
		KeepAliveConf: c.Config.ClientConfig.Keepalive,
		Timestamp: time.Now().Unix(),
		Version: version,
	}

	//buf, err := json.Marshal(res);
	buf, err := res.Marshal();
	if err != nil {
		return err
	}
	glog.V(6).Infof("postKeepAlive: %s", string(buf))
	return c.Transport.Publish("direct", "keepalives", "", buf)
}

// Deregister remove a client (or source for proxy clients) in sensu via the http api.
func (c *SensuClient) Deregister(client string) error {

	handler := "deregistration"
	if c.Config.ClientConfig.Deregistration != nil {
		handler = c.Config.ClientConfig.Deregistration.Handler
	}

	res := check.NewCheckResultFromConfig(check.CheckConfig{
		Name: "deregistration",
		Handler: &handler,
	})
	res.Output = "client initiated deregistration"
	res.Status = check.WARN
	res.Source = &client

	glog.V(6).Infof("deregistering client %s", client)
	return c.PostCheckResult(res)
}

// Marshal collects and flattens client & keepalive data and converts to []byte
func (k *KeepAliveResponse) Marshal() ([]byte, error) {
	//client info
	res := k.ClientInfo.extraFields
	if err := util.JsonStructToMap(k.ClientInfo, &res); err != nil {
		return []byte{}, err
	}

	// keepalive configs
	if err := util.JsonStructToMap(k.KeepAliveConf, &res); err != nil {
		return []byte{}, err
	}

	res["timestamp"] = k.Timestamp
	res["version"] = k.Version
	return json.Marshal(res)
}
