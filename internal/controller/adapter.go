package controller

import (
	"fmt"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	informers_v1 "k8s.io/client-go/informers/core/v1"
	informers_v1beta1 "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
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
	case "deployment":
		return &DeploymentAdapter{I: i.Extensions().V1beta1().Deployments()}
	case "pod":
		return &PodAdapter{I: i.Core().V1().Pods()}
	case "daemonset":
		return &DaemonsetAdapter{I: i.Extensions().V1beta1().DaemonSets()}
	case "service":
		return &ServiceAdapter{I: i.Core().V1().Services()}
	default:
		panic(fmt.Sprintf("'%s' is not a valid controller type", t))
	}
}

type DeploymentAdapter struct {
	I informers_v1beta1.DeploymentInformer
}

func (c *DeploymentAdapter) CheckSource(resource interface{}) string {
	r := resource.(*apps_v1.Deployment)
	return fmt.Sprintf("%s.deployment.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *DeploymentAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*apps_v1.Deployment).Annotations[CheckAnnotation]
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
	I informers_v1.PodInformer
}

func (c *PodAdapter) CheckSource(resource interface{}) string {
	r := resource.(*core_v1.Pod)
	return fmt.Sprintf("%s.pod.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *PodAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*core_v1.Pod).Annotations[CheckAnnotation]
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
	I informers_v1beta1.DaemonSetInformer
}

func (c *DaemonsetAdapter) CheckSource(resource interface{}) string {
	r := resource.(*apps_v1.DaemonSet)
	return fmt.Sprintf("%s.daemonset.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *DaemonsetAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*apps_v1.DaemonSet).Annotations[CheckAnnotation]
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

type ServiceAdapter struct {
	I informers_v1.ServiceInformer
}

func (c *ServiceAdapter) CheckSource(resource interface{}) string {
	r := resource.(*core_v1.Service)
	return fmt.Sprintf("%s.service.%s", r.ObjectMeta.Name, r.Namespace)
}
func (c *ServiceAdapter) CheckConfigs(resource interface{}) (string, bool) {
	v, ok := resource.(*core_v1.Service).Annotations[CheckAnnotation]
	return v, ok
}
func (c *ServiceAdapter) Informer() cache.SharedInformer {
	return c.I.Informer()
}
func (c *ServiceAdapter) Type() string {
	return "service"
}
func (c *ServiceAdapter) DeregisterDefault() bool {
	return false
}
