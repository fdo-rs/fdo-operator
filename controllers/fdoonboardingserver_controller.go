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

	fdov1alpha1 "github.com/empovit/fdo-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	util "github.com/redhat-cop/operator-utils/pkg/util"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	serviceInfoFilesPVC         = "fdo-serviceinfo-files-pvc"
	ownerOnboardingDefaultImage = "quay.io/vemporop/fdo-owner-onboarding-server:1.0"
	serviceInfoAPIDefaultImage  = "quay.io/vemporop/fdo-serviceinfo-api-server:1.0"
)

//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdoonboardingservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete

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
	if route, err = r.createOrUpdateRoute(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateOwnerOnboardingConfigMap(log, server, route); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateServiceInfoAPIConfigMap(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateDeployment(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateService(log, server); err != nil {
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

func (r *FDOOnboardingServerReconciler) createOrUpdateDeployment(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer) (*appsv1.Deployment, error) {

	labels := getLabels(OwnerOnboardingServiceType)
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), deploy, func() error {
		if deploy.ObjectMeta.CreationTimestamp.IsZero() {
			deploy.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
		}
		optional := false
		privileged := false
		labels := getLabels(OwnerOnboardingServiceType)
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas
		deploy.Spec.Template = corev1.PodTemplateSpec{
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
							{
								Name:      "serviceinfo-files",
								MountPath: "/etc/fdo/files",
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
					{
						Name: "serviceinfo-files",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: serviceInfoFilesPVC,
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
		}
		return ctrl.SetControllerReference(server, deploy, r.GetScheme())
	})

	if err != nil {
		log.Error(err, "Deployment reconcile failed")
		return nil, err
	} else {
		log.Info("Deployment successfully reconciled", "operation", op)
		return deploy, nil
	}
}

func (r *FDOOnboardingServerReconciler) createOrUpdateService(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer) (*corev1.Service, error) {
	labels := getLabels(OwnerOnboardingServiceType)
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), service, func() error {
		service.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       int32(8081),
					TargetPort: intstr.FromInt(8081),
				},
			},
		}
		return ctrl.SetControllerReference(server, service, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "Service reconcile failed")
		return nil, err
	} else {
		log.Info("Service successfully reconciled", "operation", op)
		return service, nil
	}
}

func (r *FDOOnboardingServerReconciler) createOrUpdateRoute(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer) (*routev1.Route, error) {
	labels := getLabels(OwnerOnboardingServiceType)
	route := &routev1.Route{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), route, func() error {
		route.Spec = routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: server.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8081),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		}
		return ctrl.SetControllerReference(server, route, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "Route reconcile failed")
		return nil, err
	} else {
		log.Info("Route successfully reconciled", "operation", op)
		return route, nil
	}
}

func (r *FDOOnboardingServerReconciler) createOrUpdateOwnerOnboardingConfigMap(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer, route *routev1.Route) (*corev1.ConfigMap, error) {
	labels := getLabels(OwnerOnboardingServiceType)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: ownerOnboardingConfigMap, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateOwnerOnboardingConfig(server, route)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"owner-onboarding-server.yml": config}
		return ctrl.SetControllerReference(server, configMap, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "ConfiMap reconcile failed")
		return nil, err
	} else {
		log.Info("ConfigMap successfully reconciled", "operation", op)
		return configMap, nil
	}
}

func (r *FDOOnboardingServerReconciler) createOrUpdateServiceInfoAPIConfigMap(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer) (*corev1.ConfigMap, error) {
	labels := getLabels(OwnerOnboardingServiceType)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: serviceInfoAPIConfigMap, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateServiceInfoAPIConfig(server)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"serviceinfo-api-server.yml": config}
		return ctrl.SetControllerReference(server, configMap, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "ConfiMap reconcile failed")
		return nil, err
	} else {
		log.Info("ConfigMap successfully reconciled", "operation", op)
		return configMap, nil
	}
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

func getLabels(svc FDOServiceType) map[string]string {
	return map[string]string{"app": "fdo", "fdo-service": string(svc)}
}
