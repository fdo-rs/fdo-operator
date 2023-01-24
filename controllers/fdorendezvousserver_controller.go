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

	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	fdov1alpha1 "github.com/empovit/fdo-operator/api/v1alpha1"
	util "github.com/redhat-cop/operator-utils/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	controllerutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	rendezvousConfigMap = "fdo-rendezvous-config"
	rendezvousImage     = "quay.io/vemporop/fdo-rendezvous-server:1.0"
)

// FDORendezvousServerReconciler reconciles a FDORendezvousServer object
type FDORendezvousServerReconciler struct {
	util.ReconcilerBase
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdorendezvousservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdorendezvousservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fdo.redhat.com,resources=fdorendezvousservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services;configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FDORendezvousServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *FDORendezvousServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("")
	log = logf.Log.WithName("fdorendezcousserver_controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling FDO rendezvous server")

	server, ok, err := r.getRendezvousServer(log, ctx, req)
	if !ok {
		return ctrl.Result{}, err
	}

	// TODO: Reload pods whenever the key/cert secrets change
	// TODO: Do we need to trigger an update of e.g. a service if there's no possibility to change it via the server CR (none of the fields affect the service)?
	// TODO: How can we deal with direct editing of the objects (ConfigMap, Route, Deployment)?
	if _, err = r.createOrUpdateConfigMap(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateDeployment(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateService(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateRoute(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	return r.ManageSuccess(ctx, server)
}

func (r *FDORendezvousServerReconciler) getRendezvousServer(log logr.Logger, ctx context.Context, req ctrl.Request) (*fdov1alpha1.FDORendezvousServer, bool, error) {
	server := &fdov1alpha1.FDORendezvousServer{}
	err := r.ReconcilerBase.GetClient().Get(ctx, req.NamespacedName, server)
	if err == nil {
		return server, true, nil
	}
	if errors.IsNotFound(err) {
		log.Info("FDORendezvousServer resource not found. Ignoring since object must have been deleted")
		return nil, false, nil
	}
	log.Error(err, "Failed to get FDORendezvousServer resource")
	return nil, false, err
}

func (r *FDORendezvousServerReconciler) createOrUpdateDeployment(log logr.Logger, server *fdov1alpha1.FDORendezvousServer) (*appsv1.Deployment, error) {
	labels := getLabels(RendezvousServiceType)
	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), deploy, func() error {
		if deploy.ObjectMeta.CreationTimestamp.IsZero() {
			deploy.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: getLabels(RendezvousServiceType),
			}
		}
		optional := false
		privileged := false
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
						Name:  "rendezvous",
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 8082,
							}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "config",
								MountPath: "/etc/fdo/rendezvous-server.conf.d",
								ReadOnly:  true,
							},
							{
								Name:      "manufacturer-cert",
								MountPath: "/etc/fdo/keys/manufacturer_cert.pem",
								SubPath:   "manufacturer_cert.pem",
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
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "config",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: rendezvousConfigMap,
								},
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

func (r *FDORendezvousServerReconciler) createOrUpdateService(log logr.Logger, server *fdov1alpha1.FDORendezvousServer) (*corev1.Service, error) {
	labels := getLabels(RendezvousServiceType)
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), service, func() error {
		service.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       int32(8082),
					TargetPort: intstr.FromInt(8082),
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

func (r *FDORendezvousServerReconciler) createOrUpdateRoute(log logr.Logger, server *fdov1alpha1.FDORendezvousServer) (*routev1.Route, error) {
	labels := getLabels(RendezvousServiceType)
	route := &routev1.Route{ObjectMeta: metav1.ObjectMeta{Name: server.Name, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), route, func() error {
		route.Spec = routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: server.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(8082),
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

func (r *FDORendezvousServerReconciler) createOrUpdateConfigMap(log logr.Logger, server *fdov1alpha1.FDORendezvousServer) (*corev1.ConfigMap, error) {
	labels := getLabels(RendezvousServiceType)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: rendezvousConfigMap, Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateConfig(server)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"rendezvous-server.yml": config}
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

func (r *FDORendezvousServerReconciler) generateConfig(fdoServer *fdov1alpha1.FDORendezvousServer) (string, error) {
	config := &RendezvousServerConfig{}
	if err := config.setValues(fdoServer); err != nil {
		return "", err
	}

	v, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FDORendezvousServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1alpha1.FDORendezvousServer{}).
		Complete(r)
}
