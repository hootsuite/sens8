
Checks Command Documentation
============================

Get latest docs via: `./sens8 -check-docs`


### `daemonset_status`

**Resources**: daemonset

Checks daemonSet pod levels via status obj given by Kubernetes. Provides full daemonSet status object in result output
Example: `daemonset_status -w 1.0 -c 0.9`

```
  -c, --crit float32   Percent of healthy (available) pods to alert critical at (default 0.9)
  -w, --warn float32   Percent of healthy (available) pods to warn at (default 1)

```


### `deployment_status`

**Resources**: deployment

Checks deployment pod levels via status obj given by Kubernetes. Provides full deployment status object in result output
Example: `deployment_status -w 0.8 -c 0.6`

```
  -c, --crit float32                    Percent of healthy (available) pods to alert critical at (default 0.8)
  -m, --min-configured-replicas int32   Alert if a deployment gets configured with a replica count below X. Often users 'suspend' a service by setting 'replicas: 0'. Intended as a simple safeguard
  -w, --warn float32                    Percent of healthy (available) pods to warn at (default 0.9)

```


### `hs_healthcheck`

**Resources**: pod, service

Make an http request to the pod or service and check the status returned in the following format: https://hootsuite.github.io/health-checks-api/#status-aggregate-get.
Example: `hs_healthcheck url=http://:::POD_IP::::8080/status/dependencies`

```
  -u, --url string   url to query. :::POD_IP::: gets replace with the pod's IP. :::HOST_IP::: gets replaced with the pod's host ip. :::CLUSTER_IP::: gets replaced by the service's ip

```


### `hs_healthcheck_v2`

**Resources**: pod, service

Make an http request to the pod or service and check the status returned in the following format: https://hootsuite.github.io/health-checks-api/#status-aggregate-get.
Example: `hs_healthcheck_v2 url=http://:::POD_IP::::8080/status/dependencies`

```
  -u, --url string   url to query. :::POD_IP::: gets replace with the pod's IP. :::HOST_IP::: gets replaced with the pod's host ip. :::CLUSTER_IP::: gets replaced by the service's ip

```


### `http`

**Resources**: pod, service

Takes a URL and checks for a 200 response (that matches a pattern, if given).
Example: Make a GET request to a pod and expect a 200 response:
`http --url=http://:::POD_IP::::8080/status/health`

```
  -d, --body string          Send a data body string with the request
  -k, --insecure             Enable insecure SSL connections
  -X, --method string        Specify a GET, POST, or PUT operation; defaults to GET (default "GET")
  -b, --response-bytes int   Print BYTES of the output (default 256)
  -O, --response-code int    Check for a specific response code (default 200)
  -u, --url string           url to query. :::POD_IP::: gets replace with the pod's IP. :::HOST_IP::: gets replaced with the pod's host ip. :::CLUSTER_IP::: gets replaced by the service's ip
  -x, --user-agent string    Specify a USER-AGENT (default "Sens8-HTTP-Check")

```



