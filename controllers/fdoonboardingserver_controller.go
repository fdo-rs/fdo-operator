/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	fdov1alpha1 "github.com/empovit/fdo-operators/api/v1alpha1"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	util "github.com/redhat-cop/operator-utils/pkg/util"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// FDOOnboardingServerReconciler reconciles a FDOOnboardingServer object
type FDOOnboardingServerReconciler struct {
	util.ReconcilerBase
	Log logr.Logger
}

type FDOServiceType string

const (
	ManufacturingServiceType   FDOServiceType = "manufacturing"
	OwnerOnboardingServiceType FDOServiceType = "owner-onboarding"
	RendezvousServiceType      FDOServiceType = "rendezvous"
)

const (
	ownerOnboardingConfigMap    = "fdo-owner-onboarding-config"
	serviceInfoAPIConfigMap     = "fdo-serviceinfo-api-config"
	ownershipVouchersPVC        = "fdo-ownership-vouchers-pvc"
	ownerOnboardingDefaultImage = "quay.io/vemporop/fdo-owner-onboarding-server:1.0"
	serviceInfoAPIDefaultImage  = "quay.io/vemporop/fdo-serviceinfo-api-server:1.0"
)

//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FDOOnboardingServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *FDOOnboardingServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("")
	log = logf.Log.WithName("fdoonboardingserver_controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling FDO onboarding server")

	// 1. Create/update deployment
	//	  a. Ensure the deployment size is the same as the spec
	// 	  b. Update the owner-onboarding image
	// 	  c. Update the serviceinfo-api image
	//	  d. Config maps are updated
	//	  e. Secrets are updated
	//    f. OV volume is updated
	// 2. Create/update onboarding service
	// 3. Create/update onboarding route
	// 4. Create/update owner-onboarding config map
	// 5. Create/update serviceinfo-api config map

	// TODO: Fix ServiceInfoAPI config structure
	// TODO: Create/update PVC or require from user?
	// TODO: How to enforce the mandatory secrets?
	// TODO: Mount a storage for files consumed by service-info
	// TODO: Allow customizing the route hostname

	server, ok, err := r.getOnboardingServer(log, ctx, req)
	if !ok {
		return ctrl.Result{}, err
	}

	r.setDefaultValues(server)
	var route *routev1.Route
	if route, err = r.getOrCreateOrUpdateRoute(log, ctx, req, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.getOrCreateOrUpdateOwnerOnboardingConfigMap(log, ctx, req, server, route); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.getOrCreateOrUpdateServiceInfoAPIConfigMap(log, ctx, req, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.getOrCreateOrUpdateDeployment(log, ctx, req, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.getOrCreateOrUpdateService(log, ctx, req, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	return r.ManageSuccess(ctx, server)
}

func (r *FDOOnboardingServerReconciler) getOnboardingServer(log logr.Logger, ctx context.Context, req ctrl.Request) (*fdov1alpha1.FDOOnboardingServer, bool, error) {
	server := &fdov1alpha1.FDOOnboardingServer{}
	err := r.ReconcilerBase.GetClient().Get(ctx, req.NamespacedName, server)
	if err == nil {
		return server, true, nil
	}
	if errors.IsNotFound(err) {
		log.Info("FDOOnboardingServer resource not found. Ignoring since object must have been deleted")
		return nil, false, nil
	}
	log.Error(err, "Failed to get FDOOnboardingServer resource")
	return nil, false, err
}

func (r *FDOOnboardingServerReconciler) setDefaultValues(server *fdov1alpha1.FDOOnboardingServer) {
	if server.Spec.OwnerOnboardingImage == "" {
		server.Spec.OwnerOnboardingImage = ownerOnboardingDefaultImage
	}
	if server.Spec.ServiceInfoImage == "" {
		server.Spec.ServiceInfoImage = serviceInfoAPIDefaultImage
	}
}

func (r *FDOOnboardingServerReconciler) getOrCreateOrUpdateDeployment(log logr.Logger, ctx context.Context, req ctrl.Request, server *fdov1alpha1.FDOOnboardingServer) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: server.Name, Namespace: server.Namespace}, deploy)
	if err == nil {
		return deploy, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Deployment")
		return nil, err
	}
	deploy = r.createDeploymentSpec(server)
	log.Info("Creating a new Deployment", "Deployment.Namespace", server.Namespace, "Deployment.Name", server.Name)
	err = r.GetClient().Create(ctx, deploy)
	if err != nil {
		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", server.Namespace, "Deployment.Name", server.Name)
		return nil, err
	}
	return deploy, nil
}

func (r *FDOOnboardingServerReconciler) getOrCreateOrUpdateService(log logr.Logger, ctx context.Context, req ctrl.Request, server *fdov1alpha1.FDOOnboardingServer) (*corev1.Service, error) {
	svc := &corev1.Service{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: server.Name, Namespace: server.Namespace}, svc)
	if err == nil {
		return svc, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Service")
		return nil, err
	}
	svc = r.createServiceSpec(server)
	log.Info("Creating a new Service", "Service.Namespace", server.Namespace, "Service.Name", server.Name)
	err = r.GetClient().Create(ctx, svc)
	if err != nil {
		log.Error(err, "Failed to create new Service", "Service.Namespace", server.Namespace, "Service.Name", server.Name)
		return nil, err
	}
	return svc, nil
}

func (r *FDOOnboardingServerReconciler) getOrCreateOrUpdateRoute(log logr.Logger, ctx context.Context, req ctrl.Request, server *fdov1alpha1.FDOOnboardingServer) (*routev1.Route, error) {
	route := &routev1.Route{}
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: server.Name, Namespace: server.Namespace}, route)
	if err == nil {
		return route, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get Route")
		return nil, err
	}
	route = r.createRouteSpec(server)
	log.Info("Creating a new Route", "Route.Namespace", server.Namespace, "Route.Name", server.Name)
	err = r.GetClient().Create(ctx, route)
	if err != nil {
		log.Error(err, "Failed to create new Route", "Route.Namespace", server.Namespace, "Route.Name", server.Name)
		return nil, err
	}
	return route, nil
}

func (r *FDOOnboardingServerReconciler) getOrCreateOrUpdateOwnerOnboardingConfigMap(log logr.Logger, ctx context.Context, req ctrl.Request, server *fdov1alpha1.FDOOnboardingServer, route *routev1.Route) (*corev1.ConfigMap, error) {
	confMap := &corev1.ConfigMap{}
	objName := ownerOnboardingConfigMap
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: objName, Namespace: server.Namespace}, confMap)
	if err == nil {
		return confMap, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get ConfigMap")
		return nil, err
	}
	ownerConf, err := r.generateOwnerOnboardingConfig(server, route)
	if err != nil {
		return nil, err
	}
	confMap = r.createOwnerOnboardingConfigMap(ownerConf, server)
	log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", server.Namespace, "ConfigMap.Name", objName)
	err = r.GetClient().Create(ctx, confMap)
	if err != nil {
		log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", server.Namespace, "ConfigMap.Name", objName)
		return nil, err
	}
	return confMap, nil
}

func (r *FDOOnboardingServerReconciler) getOrCreateOrUpdateServiceInfoAPIConfigMap(log logr.Logger, ctx context.Context, req ctrl.Request, server *fdov1alpha1.FDOOnboardingServer) (*corev1.ConfigMap, error) {
	confMap := &corev1.ConfigMap{}
	objName := serviceInfoAPIConfigMap
	err := r.GetClient().Get(ctx, types.NamespacedName{Name: objName, Namespace: server.Namespace}, confMap)
	if err == nil {
		return confMap, nil
	}
	if !errors.IsNotFound(err) {
		log.Error(err, "Failed to get ConfigMap")
		return nil, err
	}
	serviceInfoConf, err := r.generateServiceInfoAPIConfig(server)
	if err != nil {
		return nil, err
	}
	confMap = r.createServiceInfoAPIConfigMap(serviceInfoConf, server)
	log.Info("Creating a new ConfigMap", "ConfigMap.Namespace", server.Namespace, "ConfigMap.Name", objName)
	err = r.GetClient().Create(ctx, confMap)
	if err != nil {
		log.Error(err, "Failed to create new ConfigMap", "ConfigMap.Namespace", server.Namespace, "ConfigMap.Name", objName)
		return nil, err
	}
	return confMap, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FDOOnboardingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1alpha1.FDOOnboardingServer{}).
		Complete(r)
}

func (r *FDOOnboardingServerReconciler) generateOwnerOnboardingConfig(fdoServer *fdov1alpha1.FDOOnboardingServer, route *routev1.Route) (string, error) {
	config := OwnerOnboardingServerConfig{}
	if err := config.setValues(fdoServer, route); err != nil {
		return "", err
	}

	v, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (r *FDOOnboardingServerReconciler) generateServiceInfoAPIConfig(fdoServer *fdov1alpha1.FDOOnboardingServer) (string, error) {
	config := ServiceInfoAPIServerConfig{}
	if err := config.setValues(fdoServer); err != nil {
		return "", err
	}

	v, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (r *FDOOnboardingServerReconciler) createDeploymentSpec(server *fdov1alpha1.FDOOnboardingServer) *appsv1.Deployment {
	optional := false
	privileged := false
	labels := getLabels(OwnerOnboardingServiceType)
	replicas := int32(1)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: server.Spec.OwnerOnboardingImage,
							Name:  "owner-onboarding",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8081,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "owner-onboarding-config",
									MountPath: "/etc/fdo/owner-onboarding-server.conf.d",
									ReadOnly:  true,
								},
								{
									Name:      "ownership-vouchers",
									MountPath: "/etc/fdo/ownership_vouchers",
								},
								{
									Name:      "owner-cert",
									MountPath: "/etc/fdo/keys/owner_cert.pem",
									SubPath:   "owner_cert.pem",
									ReadOnly:  true,
								},
								{
									Name:      "owner-key",
									MountPath: "/etc/fdo/keys/owner_key.der",
									SubPath:   "owner_key.der",
									ReadOnly:  true,
								},
								{
									Name:      "device-ca-chain",
									MountPath: "/etc/fdo/keys/device_ca_cert.pem",
									SubPath:   "device_ca_cert.pem",
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &privileged,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
						}, {
							Image: server.Spec.ServiceInfoImage,
							Name:  "serviceinfo-api",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8083,
								}},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "serviceinfo-api-config",
									MountPath: "/etc/fdo/serviceinfo-api-server.conf.d",
									ReadOnly:  true,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &privileged,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"ALL",
									},
								},
							},
						}},
					Volumes: []corev1.Volume{
						{
							Name: "owner-onboarding-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: ownerOnboardingConfigMap,
									},
								},
							},
						},
						{
							Name: "owner-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "fdo-owner-cert",
									Items: []corev1.KeyToPath{
										{
											Key:  "owner_cert.pem",
											Path: "owner_cert.pem",
										},
									},
									Optional: &optional,
								},
							},
						},
						{
							Name: "owner-key",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "fdo-owner-key",
									Items: []corev1.KeyToPath{
										{
											Key:  "owner_key.der",
											Path: "owner_key.der",
										},
									},
									Optional: &optional,
								},
							},
						},
						{
							Name: "device-ca-chain",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "fdo-device-ca-cert",
									Items: []corev1.KeyToPath{
										{
											Key:  "device_ca_cert.pem",
											Path: "device_ca_cert.pem",
										},
									},
									Optional: &optional,
								},
							},
						},
						{
							Name: "serviceinfo-api-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: serviceInfoAPIConfigMap,
									},
								},
							},
						},
						{
							Name: "ownership-vouchers",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: ownershipVouchersPVC,
								},
							},
						},
					},
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &privileged,
						SeccompProfile: &corev1.SeccompProfile{
							Type: "RuntimeDefault",
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(server, dep, r.GetScheme())
	return dep
}

func (r *FDOOnboardingServerReconciler) createServiceSpec(server *fdov1alpha1.FDOOnboardingServer) *corev1.Service {
	labels := getLabels(OwnerOnboardingServiceType)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       int32(8081),
					TargetPort: intstr.FromInt(8081),
				},
			},
		},
	}
	ctrl.SetControllerReference(server, svc, r.GetScheme())
	return svc
}

func (r *FDOOnboardingServerReconciler) createRouteSpec(server *fdov1alpha1.FDOOnboardingServer) *routev1.Route {
	labels := getLabels(OwnerOnboardingServiceType)
	route := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
			Labels:    labels,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: server.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8081),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
	ctrl.SetControllerReference(server, route, r.GetScheme())
	return route
}

func (r *FDOOnboardingServerReconciler) createOwnerOnboardingConfigMap(config string, server *fdov1alpha1.FDOOnboardingServer) *corev1.ConfigMap {
	labels := getLabels(OwnerOnboardingServiceType)
	confMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ownerOnboardingConfigMap,
			Namespace: server.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{"owner-onboarding-server.yml": config},
	}
	ctrl.SetControllerReference(server, confMap, r.GetScheme())
	return confMap
}

func (r *FDOOnboardingServerReconciler) createServiceInfoAPIConfigMap(config string, server *fdov1alpha1.FDOOnboardingServer) *corev1.ConfigMap {
	labels := getLabels(OwnerOnboardingServiceType)
	confMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceInfoAPIConfigMap,
			Namespace: server.Namespace,
			Labels:    labels,
		},
		Data: map[string]string{"serviceinfo-api-server.yml": config},
	}
	ctrl.SetControllerReference(server, confMap, r.GetScheme())
	return confMap
}

func getLabels(svc FDOServiceType) map[string]string {
	return map[string]string{"app": "fdo", "fdo-service": string(svc)}
}
