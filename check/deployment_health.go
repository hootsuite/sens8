package check

import (
	"encoding/json"

	flag "github.com/spf13/pflag"
	"k8s.io/kubernetes/pkg/apis/extensions"
)

type DeploymentHealth struct {
	BaseCheck
	tolerance  *float32
	deployment *extensions.Deployment
}

//NewDeploymentHealth creates a new deployment health check
func NewDeploymentHealth(config CheckConfig) (Check, error) {
	dh := DeploymentHealth{}
	dh.Config = config

	// process flags
	commandLine := flag.NewFlagSet(config.Id, flag.ContinueOnError)
	dh.tolerance = commandLine.Float32("tolerance", 0.8, "health tolerance")
	err := commandLine.Parse(config.Argv[1:])
	return &dh, err
}

func (dh *DeploymentHealth) Update(resource interface{}) {
	dh.deployment = resource.(*extensions.Deployment)
}

func (dh *DeploymentHealth) Execute() (CheckResult, error) {
	res := NewCheckResultFromConfig(dh.Config)

	status := dh.deployment.Status

	res.Status = OK
	if float32(status.AvailableReplicas) / float32(status.Replicas) < *dh.tolerance {
		res.Status = CRITICAL
	} else if status.AvailableReplicas != status.Replicas {
		res.Status = WARN
	}

	o, _ := json.MarshalIndent(status, "", "  ")
	res.Output = string(o)

	return res, nil
}

// register factory
func init() {
	RegisterCheck("deployment_health", NewDeploymentHealth, []string{"deployment"})
}
