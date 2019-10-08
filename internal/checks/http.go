package checks

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/hootsuite/sens8/util"
	flag "github.com/spf13/pflag"
	core_v1 "k8s.io/api/core/v1"
)

type Http struct {
	BaseCheck
	url            *string
	method         *string
	body           *string
	ua             *string
	insecure       *bool
	response_code  *int
	response_bytes *int64
	pod            *core_v1.Pod
	resource       interface{}
	commandLine    *flag.FlagSet
	client         *http.Client
}

// insecureClient has SSL validation turned off
var insecureClient *http.Client

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureClient = &http.Client{Transport: tr}
}

//NewDeploymentStatus creates a new deployment health check
func NewHttp(config CheckConfig) (Check, error) {
	h := Http{}
	h.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	h.url = commandLine.StringP("url", "u", "", "url to query. :::POD_IP::: gets replace with the pod's IP. :::HOST_IP::: gets replaced with the pod's host ip. :::CLUSTER_IP::: gets replaced by the service's ip")
	h.method = commandLine.StringP("method", "X", "GET", "Specify a GET, POST, or PUT operation; defaults to GET")
	h.body = commandLine.StringP("body", "d", "", "Send a data body string with the request")
	h.ua = commandLine.StringP("user-agent", "x", "Sens8-HTTP-Check", "Specify a USER-AGENT")
	h.response_code = commandLine.IntP("response-code", "O", 200, "Check for a specific response code")
	h.response_bytes = commandLine.Int64P("response-bytes", "b", 256, "Print BYTES of the output")
	h.insecure = commandLine.BoolP("insecure", "k", false, "Enable insecure SSL connections")

	// @todo - implement full set of args. args should follow closely the standard check-http.rb. not all may apply
	// @todo - https://github.com/sensu-plugins/sensu-plugins-http/blob/master/bin/check-http.rb
	//h.header = commandLine.StringP("header", "H", "", "Send one or more comma-separated headers with the request")
	//h.ssl = commandLine.BoolP("ssl", "s", false, "Enabling SSL connections")
	//h.user = commandLine.StringP("username", "U", "", "A username to connect as")
	//h.password = commandLine.StringP("password", "a", "", "A password to use for the username")
	//h.cert = commandLine.StringP("cert", "c", "", "Cert FILE to use")
	//h.cacert = commandLine.StringP("cacert", "C", "", "A CA Cert FILE to use")
	//h.expiry = commandLine.IntP( "expiry", "e", 14, "Warn EXPIRE days before cert expires")
	//h.pattern = commandLine.StringP("query", "q", "", "Query for a specific PATTERN that must exist")
	//h.negpattern = commandLine.StringP("negquery", "n", "", "Query for a specific PATTERN that must be absent")
	//h.sha256checksum = commandLine.StringP("checksum", "S", "", "SHA-256 checksum")
	//h.timeout = commandLine.IntP("timeout", "t", 15, "Set the timeout in SECONDS")
	//h.redirectok = commandLine.BoolP("redirect-ok", "r", false, "Check if a redirect is ok")
	//h.redirectto = commandLine.StringP("redirect-to", "R", "", "Redirect to another page")
	//h.whole_response = commandLine.BoolP("whole-response", "w", false, "Print whole output when check fails")
	//h.require_bytes = commandLine.Int64P("require-bytes", "B", 0, "Check the response contains exactly BYTES bytes")
	//h.proxy_url = commandLine.StringP("proxy-url", "P", "", "Use a proxy server to connect")

	h.commandLine = commandLine
	if err := commandLine.Parse(config.Argv[1:]); err != nil {
		return &h, nil
	}
	if *h.url == "" {
		fmt.Errorf("--url cannot be empty")
	}

	h.client = http.DefaultClient
	if h.insecure != nil && *h.insecure == true {
		h.client = insecureClient
	}

	return &h, nil
}

func (h *Http) Usage() CheckUsage {
	d := "Takes a URL and checks for a 200 response (that matches a pattern, if given).\n"
	d += "Example: Make a GET request to a pod and expect a 200 response:\n"
	d += "`http --url=http://:::POD_IP::::8080/status/health`"
	return CheckUsage{
		Description: d,
		Flags:       h.commandLine.FlagUsages(),
	}
}

func (h *Http) Update(resource interface{}) {
	h.resource = resource
}

func (h *Http) Execute() (CheckResult, error) {
	start := time.Now()
	result := NewCheckResultFromConfig(h.Config)
	url := *h.url

	// @todo cast based on resource "apiVersion"
	t := reflect.TypeOf(h.resource).String()
	t = t[strings.LastIndex(t, ".")+1:]
	switch t {
	case "Pod":
		pod := h.resource.(*core_v1.Pod)
		url = strings.Replace(url, ":::POD_IP:::", pod.Status.PodIP, -1)
		url = strings.Replace(url, ":::HOST_IP:::", pod.Status.HostIP, -1)
	case "Service":
		service := h.resource.(*core_v1.Service)
		url = strings.Replace(url, ":::CLUSTER_IP:::", service.Spec.ClusterIP, -1)
	default:
		return result, fmt.Errorf("resource type is unknown")
	}

	// create the http request
	req, err := http.NewRequest(*h.method, url, bytes.NewBuffer([]byte(*h.body)))
	if err != nil {
		return result, err
	}
	req.Header.Add("User-Agent", *h.ua)

	// make the request
	resp, err := h.client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	result.Status = OK
	if resp.StatusCode != *h.response_code {
		result.Status = CRITICAL
	}

	// read in only the desired amount of the body to show in sensu
	o, err := ioutil.ReadAll(io.LimitReader(resp.Body, *h.response_bytes))
	if err == nil {
		result.Output += string(o)
	} else {
		result.Output += "error reading response body"
	}

	result.Duration = util.SecondsSince(start)

	return result, nil
}

// register factory
func init() {
	RegisterCheck("http", NewHttp, []string{"pod", "service"})
}
