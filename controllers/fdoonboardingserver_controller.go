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

	fdov1 "github.com/empovit/fdo-operators/api/v1"
	"github.com/go-logr/logr"
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

type serviceType string

const (
	manufacturing serviceType = "manufacturing"
	onboarding    serviceType = "owner-onboarding"
	rendezvous    serviceType = "rendezvous"
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

	server := &fdov1.FDOOnboardingServer{}
	err := r.ReconcilerBase.GetClient().Get(ctx, req.NamespacedName, server)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("FDOManufacturingServer resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get FDOManufacturingServer resource")
		return ctrl.Result{}, err
	}

	foundDep := &appsv1.Deployment{}
	err = r.GetClient().Get(ctx, types.NamespacedName{Name: server.Name, Namespace: server.Namespace}, foundDep)
	if err != nil && errors.IsNotFound(err) {
		dep := r.createDeploymentSpec(server)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.GetClient().Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// 1. Create/update deployment
	//	  a. Ensure the deployment size is the same as the spec
	// 	  b. Update the owner-onboarding image
	// 	  c. Update the serviceinfo-api image
	// 2. Create/update onboarding service
	// 3. Create/update onboarding route
	// 4. Create/update owner-onboarding config map
	// 5. Create/update serviceinfo-api config map

	ownerConf, err := r.generateOwnerOnboardingConfig(server)
	if err != nil {
		return r.ManageError(ctx, server, err)
	}

	log.Info(ownerConf)

	serviceInfoConf, err := r.generateServiceInfoAPIConfig(server)
	if err != nil {
		return r.ManageError(ctx, server, err)
	}

	log.Info(serviceInfoConf)

	return r.ManageSuccess(ctx, server)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FDOOnboardingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1.FDOOnboardingServer{}).
		Complete(r)
}

func (r *FDOOnboardingServerReconciler) generateOwnerOnboardingConfig(fdoServer *fdov1.FDOOnboardingServer) (string, error) {
	config := OwnerOnboardingServerConfig{}
	if err := config.setValues(fdoServer); err != nil {
		return "", err
	}

	v, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (r *FDOOnboardingServerReconciler) generateServiceInfoAPIConfig(fdoServer *fdov1.FDOOnboardingServer) (string, error) {
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

func (r *FDOOnboardingServerReconciler) createDeploymentSpec(server *fdov1.FDOOnboardingServer) *appsv1.Deployment {
	optional := false
	labels := getLabels(onboarding)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &server.Spec.Replicas,
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
						}},
					Volumes: []corev1.Volume{
						{
							Name: "owner-onboarding-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "fdo-owner-onboarding-config",
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
										Name: "fdo-serviceinfo-api-config",
									},
								},
							},
						},
						{
							Name: "ownership-vouchers",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "fdo-ownership-vouchers-pvc",
								},
							},
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(server, dep, r.GetScheme())
	return dep
}

func (r *FDOOnboardingServerReconciler) createServiceSpec(server *fdov1.FDOOnboardingServer) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: server.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: getLabels(onboarding),
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

func getLabels(svc serviceType) map[string]string {
	return map[string]string{"app": "fdo", "service": string(svc)}
}
