/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trait

import (
	"fmt"
	"path"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	serving "knative.dev/serving/pkg/apis/serving/v1"

	v1 "github.com/apache/camel-k/pkg/apis/camel/v1"
	"github.com/apache/camel-k/pkg/util"
	"github.com/apache/camel-k/pkg/util/camel"
	"github.com/apache/camel-k/pkg/util/defaults"
	"github.com/apache/camel-k/pkg/util/envvar"
	"github.com/apache/camel-k/pkg/util/kubernetes"
)

const (
	defaultContainerName     = "integration"
	defaultContainerPort     = 8080
	defaultContainerPortName = "http"
	defaultServicePort       = 80
	containerTraitID         = "container"
)

// The Container trait can be used to configure properties of the container where the integration will run.
//
// It also provides configuration for Services associated to the container.
//
// +camel-k:trait=container
// nolint: tagliatelle
type containerTrait struct {
	BaseTrait `property:",squash"`

	Auto *bool `property:"auto" json:"auto,omitempty"`

	// The minimum amount of CPU required.
	RequestCPU string `property:"request-cpu" json:"requestCPU,omitempty"`
	// The minimum amount of memory required.
	RequestMemory string `property:"request-memory" json:"requestMemory,omitempty"`
	// The maximum amount of CPU required.
	LimitCPU string `property:"limit-cpu" json:"limitCPU,omitempty"`
	// The maximum amount of memory required.
	LimitMemory string `property:"limit-memory" json:"limitMemory,omitempty"`

	// Can be used to enable/disable exposure via kubernetes Service.
	Expose *bool `property:"expose" json:"expose,omitempty"`
	// To configure a different port exposed by the container (default `8080`).
	Port int `property:"port" json:"port,omitempty"`
	// To configure a different port name for the port exposed by the container. It defaults to `http` only when the `expose` parameter is true.
	PortName string `property:"port-name" json:"portName,omitempty"`
	// To configure under which service port the container port is to be exposed (default `80`).
	ServicePort int `property:"service-port" json:"servicePort,omitempty"`
	// To configure under which service port name the container port is to be exposed (default `http`).
	ServicePortName string `property:"service-port-name" json:"servicePortName,omitempty"`
	// The main container name. It's named `integration` by default.
	Name string `property:"name" json:"name,omitempty"`
	// The main container image
	Image string `property:"image" json:"image,omitempty"`
	// The pull policy: Always|Never|IfNotPresent
	ImagePullPolicy corev1.PullPolicy `property:"image-pull-policy" json:"imagePullPolicy,omitempty"`

	// DeprecatedProbesEnabled enable/disable probes on the container (default `false`)
	// Deprecated: replaced by the health trait.
	DeprecatedProbesEnabled *bool `property:"probes-enabled" json:"probesEnabled,omitempty"`
	// Scheme to use when connecting. Defaults to HTTP. Applies to the liveness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessScheme string `property:"liveness-scheme" json:"livenessScheme,omitempty"`
	// Number of seconds after the container has started before liveness probes are initiated.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessInitialDelay int32 `property:"liveness-initial-delay" json:"livenessInitialDelay,omitempty"`
	// Number of seconds after which the probe times out. Applies to the liveness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessTimeout int32 `property:"liveness-timeout" json:"livenessTimeout,omitempty"`
	// How often to perform the probe. Applies to the liveness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessPeriod int32 `property:"liveness-period" json:"livenessPeriod,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Applies to the liveness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessSuccessThreshold int32 `property:"liveness-success-threshold" json:"livenessSuccessThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Applies to the liveness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedLivenessFailureThreshold int32 `property:"liveness-failure-threshold" json:"livenessFailureThreshold,omitempty"`
	// Scheme to use when connecting. Defaults to HTTP. Applies to the readiness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessScheme string `property:"readiness-scheme" json:"readinessScheme,omitempty"`
	// Number of seconds after the container has started before readiness probes are initiated.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessInitialDelay int32 `property:"readiness-initial-delay" json:"readinessInitialDelay,omitempty"`
	// Number of seconds after which the probe times out. Applies to the readiness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessTimeout int32 `property:"readiness-timeout" json:"readinessTimeout,omitempty"`
	// How often to perform the probe. Applies to the readiness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessPeriod int32 `property:"readiness-period" json:"readinessPeriod,omitempty"`
	// Minimum consecutive successes for the probe to be considered successful after having failed.
	// Applies to the readiness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessSuccessThreshold int32 `property:"readiness-success-threshold" json:"readinessSuccessThreshold,omitempty"`
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	// Applies to the readiness probe.
	// Deprecated: replaced by the health trait.
	DeprecatedReadinessFailureThreshold int32 `property:"readiness-failure-threshold" json:"readinessFailureThreshold,omitempty"`
}

func newContainerTrait() Trait {
	return &containerTrait{
		BaseTrait:                 NewBaseTrait(containerTraitID, 1600),
		Port:                      defaultContainerPort,
		ServicePort:               defaultServicePort,
		ServicePortName:           defaultContainerPortName,
		Name:                      defaultContainerName,
		DeprecatedProbesEnabled:   pointer.Bool(false),
		DeprecatedLivenessScheme:  string(corev1.URISchemeHTTP),
		DeprecatedReadinessScheme: string(corev1.URISchemeHTTP),
	}
}

func (t *containerTrait) Configure(e *Environment) (bool, error) {
	if !pointer.BoolDeref(t.Enabled, true) {
		return false, nil
	}

	if !e.IntegrationInPhase(v1.IntegrationPhaseInitialization) && !e.IntegrationInRunningPhases() {
		return false, nil
	}

	if pointer.BoolDeref(t.Auto, true) {
		if t.Expose == nil {
			e := e.Resources.GetServiceForIntegration(e.Integration) != nil
			t.Expose = &e
		}
	}

	if !isValidPullPolicy(t.ImagePullPolicy) {
		return false, fmt.Errorf("unsupported pull policy %s", t.ImagePullPolicy)
	}

	return true, nil
}

func isValidPullPolicy(policy corev1.PullPolicy) bool {
	return policy == "" || policy == corev1.PullAlways || policy == corev1.PullIfNotPresent || policy == corev1.PullNever
}

func (t *containerTrait) Apply(e *Environment) error {
	if err := t.configureImageIntegrationKit(e); err != nil {
		return err
	}
	return t.configureContainer(e)
}

// IsPlatformTrait overrides base class method.
func (t *containerTrait) IsPlatformTrait() bool {
	return true
}

func (t *containerTrait) configureImageIntegrationKit(e *Environment) error {
	if t.Image != "" {
		if e.Integration.Spec.IntegrationKit != nil {
			return fmt.Errorf(
				"unsupported configuration: a container image has been set in conjunction with an IntegrationKit %v",
				e.Integration.Spec.IntegrationKit)
		}

		kitName := fmt.Sprintf("kit-%s", e.Integration.Name)
		kit := v1.NewIntegrationKit(e.Integration.Namespace, kitName)
		kit.Spec.Image = t.Image

		// Add some information for post-processing, this may need to be refactored
		// to a proper data structure
		kit.Labels = map[string]string{
			v1.IntegrationKitTypeLabel:            v1.IntegrationKitTypeExternal,
			kubernetes.CamelCreatorLabelKind:      v1.IntegrationKind,
			kubernetes.CamelCreatorLabelName:      e.Integration.Name,
			kubernetes.CamelCreatorLabelNamespace: e.Integration.Namespace,
			kubernetes.CamelCreatorLabelVersion:   e.Integration.ResourceVersion,
		}

		if kit.Annotations == nil {
			kit.Annotations = make(map[string]string)
		}
		if v, ok := e.Integration.Annotations[v1.PlatformSelectorAnnotation]; ok {
			kit.Annotations[v1.PlatformSelectorAnnotation] = v
		}
		operatorID := defaults.OperatorID()
		if operatorID != "" {
			kit.Annotations[v1.OperatorIDAnnotation] = operatorID
		}

		t.L.Infof("image %s", kit.Spec.Image)
		e.Resources.Add(kit)
		e.Integration.SetIntegrationKit(kit)
	}
	return nil
}

func (t *containerTrait) configureContainer(e *Environment) error {
	if e.ApplicationProperties == nil {
		e.ApplicationProperties = make(map[string]string)
	}

	container := corev1.Container{
		Name:  t.Name,
		Image: e.Integration.Status.Image,
		Env:   make([]corev1.EnvVar, 0),
	}

	if t.ImagePullPolicy != "" {
		container.ImagePullPolicy = t.ImagePullPolicy
	}

	// combine Environment of integration with platform, kit, integration
	for _, env := range e.collectConfigurationPairs("env") {
		envvar.SetVal(&container.Env, env.Name, env.Value)
	}

	envvar.SetVal(&container.Env, "CAMEL_K_DIGEST", e.Integration.Status.Digest)
	envvar.SetVal(&container.Env, "CAMEL_K_CONF", path.Join(camel.BasePath, "application.properties"))
	envvar.SetVal(&container.Env, "CAMEL_K_CONF_D", camel.ConfDPath)

	e.addSourcesProperties()
	if props, err := e.computeApplicationProperties(); err != nil {
		return err
	} else if props != nil {
		e.Resources.Add(props)
	}

	t.configureResources(e, &container)
	if pointer.BoolDeref(t.Expose, false) {
		t.configureService(e, &container)
	}
	t.configureCapabilities(e)

	var containers *[]corev1.Container
	visited := false

	// Deployment
	if err := e.Resources.VisitDeploymentE(func(deployment *appsv1.Deployment) error {
		for _, envVar := range e.EnvVars {
			envvar.SetVar(&container.Env, envVar)
		}

		containers = &deployment.Spec.Template.Spec.Containers
		visited = true
		return nil
	}); err != nil {
		return err
	}

	// Knative Service
	if err := e.Resources.VisitKnativeServiceE(func(service *serving.Service) error {
		for _, env := range e.EnvVars {
			switch {
			case env.ValueFrom == nil:
				envvar.SetVar(&container.Env, env)
			case env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath == "metadata.namespace":
				envvar.SetVar(&container.Env, corev1.EnvVar{Name: env.Name, Value: e.Integration.Namespace})
			case env.ValueFrom.FieldRef != nil:
				t.L.Infof("Skipping environment variable %s (fieldRef)", env.Name)
			case env.ValueFrom.ResourceFieldRef != nil:
				t.L.Infof("Skipping environment variable %s (resourceFieldRef)", env.Name)
			default:
				envvar.SetVar(&container.Env, env)
			}
		}

		containers = &service.Spec.ConfigurationSpec.Template.Spec.Containers
		visited = true
		return nil
	}); err != nil {
		return err
	}

	// CronJob
	if err := e.Resources.VisitCronJobE(func(cron *v1beta1.CronJob) error {
		for _, envVar := range e.EnvVars {
			envvar.SetVar(&container.Env, envVar)
		}

		containers = &cron.Spec.JobTemplate.Spec.Template.Spec.Containers
		visited = true
		return nil
	}); err != nil {
		return err
	}

	if visited {
		*containers = append(*containers, container)
	}

	return nil
}

func (t *containerTrait) configureService(e *Environment, container *corev1.Container) {
	service := e.Resources.GetServiceForIntegration(e.Integration)
	if service == nil {
		return
	}

	name := t.PortName
	if name == "" {
		name = defaultContainerPortName
	}

	containerPort := corev1.ContainerPort{
		Name:          name,
		ContainerPort: int32(t.Port),
		Protocol:      corev1.ProtocolTCP,
	}

	servicePort := corev1.ServicePort{
		Name:       t.ServicePortName,
		Port:       int32(t.ServicePort),
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromString(name),
	}

	e.Integration.Status.SetCondition(
		v1.IntegrationConditionServiceAvailable,
		corev1.ConditionTrue,
		v1.IntegrationConditionServiceAvailableReason,

		// service -> container
		fmt.Sprintf("%s(%s/%d) -> %s(%s/%d)",
			service.Name, servicePort.Name, servicePort.Port,
			container.Name, containerPort.Name, containerPort.ContainerPort),
	)

	container.Ports = append(container.Ports, containerPort)
	service.Spec.Ports = append(service.Spec.Ports, servicePort)

	// Mark the service as a user service
	service.Labels["camel.apache.org/service.type"] = v1.ServiceTypeUser
}

func (t *containerTrait) configureResources(_ *Environment, container *corev1.Container) {
	// Requests
	if container.Resources.Requests == nil {
		container.Resources.Requests = make(corev1.ResourceList)
	}

	if t.RequestCPU != "" {
		v, err := resource.ParseQuantity(t.RequestCPU)
		if err != nil {
			t.L.Error(err, "unable to parse quantity", "request-cpu", t.RequestCPU)
		} else {
			container.Resources.Requests[corev1.ResourceCPU] = v
		}
	}
	if t.RequestMemory != "" {
		v, err := resource.ParseQuantity(t.RequestMemory)
		if err != nil {
			t.L.Error(err, "unable to parse quantity", "request-memory", t.RequestMemory)
		} else {
			container.Resources.Requests[corev1.ResourceMemory] = v
		}
	}

	// Limits
	if container.Resources.Limits == nil {
		container.Resources.Limits = make(corev1.ResourceList)
	}

	if t.LimitCPU != "" {
		v, err := resource.ParseQuantity(t.LimitCPU)
		if err != nil {
			t.L.Error(err, "unable to parse quantity", "limit-cpu", t.LimitCPU)
		} else {
			container.Resources.Limits[corev1.ResourceCPU] = v
		}
	}
	if t.LimitMemory != "" {
		v, err := resource.ParseQuantity(t.LimitMemory)
		if err != nil {
			t.L.Error(err, "unable to parse quantity", "limit-memory", t.LimitMemory)
		} else {
			container.Resources.Limits[corev1.ResourceMemory] = v
		}
	}
}

func (t *containerTrait) configureCapabilities(e *Environment) {
	if util.StringSliceExists(e.Integration.Status.Capabilities, v1.CapabilityRest) {
		e.ApplicationProperties["camel.context.rest-configuration.component"] = "platform-http"
	}
}
