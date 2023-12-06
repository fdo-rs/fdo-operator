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
	"fmt"

	fdov1alpha1 "github.com/fdo-rs/fdo-operator/api/v1alpha1"
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

// FDOManufacturingServerReconciler reconciles a FDOManufacturingServer object
type FDOManufacturingServerReconciler struct {
	util.ReconcilerBase
	Log logr.Logger
}

const (
	manufacturingConfigMapTemplate = "%s-config"
	manufacturingDefaultImage      = "quay.io/vemporop/fdo-manufacturing-server:rhel9.3"
)

//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdomanufacturingservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdomanufacturingservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdomanufacturingservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FDOManufacturingServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FDOManufacturingServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("")
	log = logf.Log.WithName("fdomanufacturingserver_controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling FDO manufacturing server")

	server, ok, err := r.getManufacturingServer(log, ctx, req)
	if !ok {
		return ctrl.Result{}, err
	}

	r.setDefaultValues(server)

	if _, err = r.createOrUpdateRoute(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateConfigMap(log, server); err != nil {
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

func (r *FDOManufacturingServerReconciler) getManufacturingServer(log logr.Logger, ctx context.Context, req ctrl.Request) (*fdov1alpha1.FDOManufacturingServer, bool, error) {
	server := &fdov1alpha1.FDOManufacturingServer{}
	err := r.ReconcilerBase.GetClient().Get(ctx, req.NamespacedName, server)
	if err == nil {
		return server, true, nil
	}
	if errors.IsNotFound(err) {
		log.Info("FDOManufacturingServer resource not found. Ignoring since object must have been deleted")
		return nil, false, nil
	}
	log.Error(err, "Failed to get FDOManufacturingServer resource")
	return nil, false, err
}

func (r *FDOManufacturingServerReconciler) setDefaultValues(server *fdov1alpha1.FDOManufacturingServer) {
	if server.Spec.Image == "" {
		server.Spec.Image = manufacturingDefaultImage
	}
}

func (r *FDOManufacturingServerReconciler) createOrUpdateDeployment(log logr.Logger, server *fdov1alpha1.FDOManufacturingServer) (*appsv1.Deployment, error) {

	labels := getLabels(ManufacturingServiceType)
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), deploy, func() error {
		if deploy.ObjectMeta.CreationTimestamp.IsZero() {
			deploy.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}
		}
		optional := false
		privilegeEscalation := false
		nonRoot := true
		labels := getLabels(ManufacturingServiceType)
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas
		deploy.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Image: server.Spec.Image,
						Name:  "manufacturing",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8080,
							}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "manufacturing-config",
								MountPath: "/etc/fdo/manufacturing-server.conf.d",
								ReadOnly:  true,
							},
							{
								Name:      "ownership-vouchers",
								MountPath: "/etc/fdo/ownership_vouchers",
							},
							{
								Name:      "diun-cert",
								MountPath: "/etc/fdo/keys/diun_cert.pem",
								SubPath:   "diun_cert.pem",
								ReadOnly:  true,
							},
							{
								Name:      "diun-key",
								MountPath: "/etc/fdo/keys/diun_key.der",
								SubPath:   "diun_key.der",
								ReadOnly:  true,
							},
							{
								Name:      "manufacturer-cert",
								MountPath: "/etc/fdo/keys/manufacturer_cert.pem",
								SubPath:   "manufacturer_cert.pem",
								ReadOnly:  true,
							},
							{
								Name:      "manufacturer-key",
								MountPath: "/etc/fdo/keys/manufacturer_key.der",
								SubPath:   "manufacturer_key.der",
								ReadOnly:  true,
							},
							{
								Name:      "owner-cert",
								MountPath: "/etc/fdo/keys/owner_cert.pem",
								SubPath:   "owner_cert.pem",
								ReadOnly:  true,
							},
							{
								Name:      "device-ca-key",
								MountPath: "/etc/fdo/keys/device_ca_key.der",
								SubPath:   "device_ca_key.der",
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
							AllowPrivilegeEscalation: &privilegeEscalation,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
					}},
				Volumes: []corev1.Volume{
					{
						Name: "manufacturing-config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf(manufacturingConfigMapTemplate, server.Name),
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
						Name: "diun-cert",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "fdo-diun-cert",
								Items: []corev1.KeyToPath{
									{
										Key:  "diun_cert.pem",
										Path: "diun_cert.pem",
									},
								},
								Optional: &optional,
							},
						},
					},
					{
						Name: "diun-key",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "fdo-diun-key",
								Items: []corev1.KeyToPath{
									{
										Key:  "diun_key.der",
										Path: "diun_key.der",
									},
								},
								Optional: &optional,
							},
						},
					},
					{
						Name: "manufacturer-cert",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "fdo-manufacturer-cert",
								Items: []corev1.KeyToPath{
									{
										Key:  "manufacturer_cert.pem",
										Path: "manufacturer_cert.pem",
									},
								},
								Optional: &optional,
							},
						},
					},
					{
						Name: "manufacturer-key",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "fdo-manufacturer-key",
								Items: []corev1.KeyToPath{
									{
										Key:  "manufacturer_key.der",
										Path: "manufacturer_key.der",
									},
								},
								Optional: &optional,
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
						Name: "device-ca-key",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "fdo-device-ca-key",
								Items: []corev1.KeyToPath{
									{
										Key:  "device_ca_key.der",
										Path: "device_ca_key.der",
									},
								},
								Optional: &optional,
							},
						},
					},
				},
				SecurityContext: &corev1.PodSecurityContext{
					RunAsNonRoot: &nonRoot,
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

func (r *FDOManufacturingServerReconciler) createOrUpdateService(log logr.Logger, server *fdov1alpha1.FDOManufacturingServer) (*corev1.Service, error) {
	labels := getLabels(ManufacturingServiceType)
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), service, func() error {
		service.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       int32(8080),
					TargetPort: intstr.FromInt(8080),
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

func (r *FDOManufacturingServerReconciler) createOrUpdateRoute(log logr.Logger, server *fdov1alpha1.FDOManufacturingServer) (*routev1.Route, error) {
	labels := getLabels(ManufacturingServiceType)
	route := &routev1.Route{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), route, func() error {
		route.Spec = routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: server.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8080),
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

func (r *FDOManufacturingServerReconciler) createOrUpdateConfigMap(log logr.Logger, server *fdov1alpha1.FDOManufacturingServer) (*corev1.ConfigMap, error) {
	labels := getLabels(ManufacturingServiceType)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(manufacturingConfigMapTemplate, server.Name), Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateConfig(server)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"manufacturing-server.yml": config}
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
func (r *FDOManufacturingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1alpha1.FDOManufacturingServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Complete(r)
}

func (r *FDOManufacturingServerReconciler) generateConfig(fdoServer *fdov1alpha1.FDOManufacturingServer) (string, error) {
	config := ManufacturingServerConfig{}
	if err := config.setValues(fdoServer); err != nil {
		return "", err
	}

	v, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(v), nil
}
