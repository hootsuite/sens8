package controller

import (
	"fmt"
	"k8s.io/kubernetes/pkg/controller/informers"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/api"
)

type ResourceAdapter interface {
	CheckSource(resource interface{}) string
	CheckConfigs(resource interface{}) (string, bool)
	Informer() cache.SharedInformer
	Type() string
	DeregisterDefault() bool
}

// ResourceAdapterFactory creates an adapter for the given resource type (t) and informer factory
func ResourceAdapterFactory(t string, i informers.SharedInformerFactory) ResourceAdapter {
	switch t {
	case "deployment": return &DeploymentAdapter{I:i.Deployments()}
	case "pod": return &PodAdapter{I: i.Pods()}
	}
	return nil
}


type DeploymentAdapter struct {
	I informers.DeploymentInformer
}
func (c *DeploymentAdapter) CheckSource(resource interface{}) string {
	r := resource.(*extensions.Deployment)
	return fmt.Sprintf("%s.deployment.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *DeploymentAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*extensions.Deployment).Annotations[CheckAnnotation]
	return v, ok
}
func (c *DeploymentAdapter) Informer() cache.SharedInformer {
	return c.I.Informer()
}
func (c *DeploymentAdapter) Type() string {
	return "deployment"
}
func (c *DeploymentAdapter) DeregisterDefault() bool {
	return false
}


type PodAdapter struct {
	I informers.PodInformer
}
func (c *PodAdapter) CheckSource(resource interface{}) string {
	r := resource.(*api.Pod)
	return fmt.Sprintf("%s.pod.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *PodAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*api.Pod).Annotations[CheckAnnotation]
	return v, ok
}
func (c *PodAdapter) Informer() cache.SharedInformer {
	return c.I.Informer()
}
func (c *PodAdapter) Type() string {
	return "pod"
}
func (c *PodAdapter) DeregisterDefault() bool {
	return true
}

