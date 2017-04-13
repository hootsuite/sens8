Checks Commands
===============

Get latest docs via: `./sens8 -check-commands`

### `deployment_status`

**Resources**: deployment

Checks deployment pod levels via status obj given by Kubernetes. Provides full deployment status object in result output
Example: `deployment_status -w 0.8 -c 0.6`

```
  -c, --crit value                      Percent of healthy pods to alert critical at (default 0.8)
  -m, --min-configured-replicas value   Alert if a deployment gets configured with a replica count below X. Often users 'suspend' a service by setting 'replicas: 0'. Intended as a simple safeguard
  -w, --warn value                      Percent of healthy pods to warn at (default 0.9)

```

### `hs_healthcheck`

**Resources**: pod, service

Make an http request to the pod or service and check the status returned in the following format: https://hootsuite.github.io/health-checks-api/#status-aggregate-get.
Example: `hs_healthcheck url=http://:::POD_IP::::8080/status/dependencies`

```
  -u, --url string   url to query. :::POD_IP::: gets replace with the pod's IP. :::HOST_IP::: gets replaced with the pod's host ip. :::CUSTER_IP::: gets replaced by the service's ip

```

