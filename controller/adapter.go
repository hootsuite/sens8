package controller

import (
	"fmt"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/pkg/apis/extensions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/pkg/api"
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
	case "deployment": return &DeploymentAdapter{I:i.Extensions().V1beta1().Deployments()}
	case "pod": return &PodAdapter{I: i.Core().V1().Pods()}
	}
	return nil
}


type DeploymentAdapter struct {
	I v1beta1.DeploymentInformer
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
	I v1.PodInformer
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


type DaemonsetAdapter struct {
	I v1beta1.DaemonSetInformer
}
func (c *DaemonsetAdapter) CheckSource(resource interface{}) string {
	r := resource.(*extensions.DaemonSet)
	return fmt.Sprintf("%s.pod.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *DaemonsetAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*extensions.DaemonSet).Annotations[CheckAnnotation]
	return v, ok
}
func (c *DaemonsetAdapter) Informer() cache.SharedInformer {
	return c.I.Informer()
}
func (c *DaemonsetAdapter) Type() string {
	return "daemonset"
}
func (c *DaemonsetAdapter) DeregisterDefault() bool {
	return false
}
