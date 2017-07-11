package rabbitmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var errNoUserInURI = errors.New("No user found in URI")

type TransportConfig struct {
	Host      string      `json:"host,omitempty"`
	Port      json.Number `json:"port,omitempty"`
	Vhost     string      `json:"vhost,omitempty"`
	User      string      `json:"user,omitempty"`
	Password  string      `json:"password,omitempty"`
	Heartbeat json.Number `json:"heartbeat,omitempty"`
	Prefetch  json.Number `json:"prefetch,omitempty"`
	Ssl       struct {
		CertChainFile  string `json:"cert_chain_file,omitempty"`
		PrivateKeyFile string `json:"private_key_file,omitempty"`
	} `json:"ssl,omitempty"`
}

func (c *TransportConfig) GetURI() string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/%s",
		c.User,
		c.Password,
		c.Host,
		c.Port.String(),
		url.QueryEscape(c.Vhost),
	)
}

func NewTransportConfig(uri string) (*TransportConfig, error) {
	uriComponents, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse the URI: %s", err)
	}

	if !strings.Contains(uriComponents.Host, ":") {
		return nil, fmt.Errorf("Failed to determine the port for host: %s", uriComponents.Host)
	}

	host, port, err := net.SplitHostPort(uriComponents.Host)
	if err != nil {
		return nil, fmt.Errorf("Failed to separate the host name from the port: %s", err)
	}

	if uriComponents.User == nil {
		return nil, errNoUserInURI
	}

	user := uriComponents.User.Username()
	password, _ := uriComponents.User.Password()

	return &TransportConfig{
		Host:     host,
		Port:     json.Number(port),
		Vhost:    uriComponents.Path[1:], // Discard the leading slash
		User:     user,
		Password: password,
	}, nil
}
