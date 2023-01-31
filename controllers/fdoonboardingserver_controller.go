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
	"sort"
	"time"

	fdov1alpha1 "github.com/empovit/fdo-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	util "github.com/redhat-cop/operator-utils/pkg/util"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	ownerOnboardingConfigMapTemplate = "%s-owner-onboarding-config"
	serviceInfoAPIConfigMapTemplate  = "%s-serviceinfo-api-config"
	ownershipVouchersPVC             = "fdo-ownership-vouchers-pvc"
	serviceInfoFilesPVC              = "fdo-serviceinfo-files-pvc"
	ownerOnboardingDefaultImage      = "quay.io/vemporop/fdo-owner-onboarding-server:1.0"
	serviceInfoAPIDefaultImage       = "quay.io/vemporop/fdo-serviceinfo-api-server:1.0"
)

const (
	FileKey          = "fdo.serviceinfo.file/name"
	PathKey          = "fdo.serviceinfo.file/path"
	PermissionsKey   = "fdo.serviceinfo.file/permissions"
	FilePathTemplate = "/etc/fdo/files/%s/%s"
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *FDOOnboardingServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("")
	log = logf.Log.WithName("fdoonboardingserver_controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling FDO onboarding server")

	// TODO:
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

	files, err := r.listConfigMaps(log, ctx, req, server.Name)
	if err != nil {
		return r.ManageErrorWithRequeue(ctx, server, err, 30*time.Second) // allow time for the user to fix the configuration
	}

	if _, err = r.createOrUpdateServiceInfoAPIConfigMap(log, server, files); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateDeployment(log, server, files); err != nil {
		return r.ManageError(ctx, server, err)
	}

	if _, err = r.createOrUpdateService(log, server); err != nil {
		return r.ManageError(ctx, server, err)
	}

	// Allow the controller to pick up new serviceinfo files
	return r.ManageSuccessWithRequeue(ctx, server, 5*time.Minute)
}

// SetupWithManager sets up the controller with the Manager.
func (r *FDOOnboardingServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1alpha1.FDOOnboardingServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Complete(r)
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

func (r *FDOOnboardingServerReconciler) createOrUpdateDeployment(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer, files []ServiceInfoFile) (*appsv1.Deployment, error) {

	labels := getLabels(OwnerOnboardingServiceType)
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
		labels := getLabels(OwnerOnboardingServiceType)
		replicas := int32(1)
		deploy.Spec.Replicas = &replicas

		serviceInfoVolumeMounts := []corev1.VolumeMount{
			{
				Name:      "serviceinfo-api-config",
				MountPath: "/etc/fdo/serviceinfo-api-server.conf.d",
				ReadOnly:  true,
			},
		}
		volumes := []corev1.Volume{
			{
				Name: "owner-onboarding-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf(ownerOnboardingConfigMapTemplate, server.Name),
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
							Name: fmt.Sprintf(serviceInfoAPIConfigMapTemplate, server.Name),
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
		}

		for _, f := range files {
			serviceInfoVolumeMounts = append(serviceInfoVolumeMounts, corev1.VolumeMount{
				Name:      f.ConfigMap,
				MountPath: fmt.Sprintf("/etc/fdo/files/%s", f.ConfigMap),
				ReadOnly:  true,
			})
			volumes = append(volumes, corev1.Volume{
				Name: f.ConfigMap,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: f.ConfigMap,
						},
					},
				},
			})
		}

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
							AllowPrivilegeEscalation: &privilegeEscalation,
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
						VolumeMounts: serviceInfoVolumeMounts,
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &privilegeEscalation,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									"ALL",
								},
							},
						},
					}},
				Volumes: volumes,
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
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(ownerOnboardingConfigMapTemplate, server.Name), Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateOwnerOnboardingConfig(server, route)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"owner-onboarding-server.yml": config}
		return ctrl.SetControllerReference(server, configMap, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "ConfigMap reconcile failed for owner-onboarding")
		return nil, err
	} else {
		log.Info("ConfigMap successfully reconciled for owner-onboarding", "operation", op)
		return configMap, nil
	}
}

func (r *FDOOnboardingServerReconciler) createOrUpdateServiceInfoAPIConfigMap(log logr.Logger, server *fdov1alpha1.FDOOnboardingServer, files []ServiceInfoFile) (*corev1.ConfigMap, error) {
	labels := getLabels(OwnerOnboardingServiceType)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf(serviceInfoAPIConfigMapTemplate, server.Name), Namespace: server.Namespace, Labels: labels}}
	op, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), configMap, func() error {
		config, err := r.generateServiceInfoAPIConfig(server, files)
		if err != nil {
			return err
		}
		configMap.Data = map[string]string{"serviceinfo-api-server.yml": config}
		return ctrl.SetControllerReference(server, configMap, r.GetScheme())
	})
	if err != nil {
		log.Error(err, "ConfigMap reconcile failed for serviceinfo-api")
		return nil, err
	} else {
		log.Info("ConfigMap successfully reconciled for serviceinfo-api", "operation", op)
		return configMap, nil
	}
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

func (r *FDOOnboardingServerReconciler) generateServiceInfoAPIConfig(fdoServer *fdov1alpha1.FDOOnboardingServer, files []ServiceInfoFile) (string, error) {
	config := ServiceInfoAPIServerConfig{}
	if err := config.setValues(fdoServer, files); err != nil {
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

func (r *FDOOnboardingServerReconciler) listConfigMaps(log logr.Logger, ctx context.Context, req ctrl.Request, name string) ([]ServiceInfoFile, error) {
	require, err := labels.NewRequirement("serviceinfo.file/owner", selection.Equals, []string{name})
	if err != nil {
		return nil, err
	}
	c := r.ReconcilerBase.GetClient()
	selector := labels.NewSelector()
	selector = selector.Add(*require)
	foundCms := &corev1.ConfigMapList{}
	if err := c.List(ctx, foundCms, &client.ListOptions{
		Namespace:     req.Namespace,
		LabelSelector: selector,
	}); err != nil {
		return nil, err
	}

	files := make([]ServiceInfoFile, len(foundCms.Items))
	for i, cm := range foundCms.Items {
		config := &ServiceInfoFile{}
		if err := readServiceInfoFileFromConfigMap(cm, config); err != nil {
			return nil, err
		}
		log.Info("ServiceInfo file found", "name", cm.Name, "namespace", cm.Namespace, "config", config)
		files[i] = *config
	}

	// maintain stable order to prevent unnecessary updates
	sort.SliceStable(files, func(i, j int) bool { return files[i].ConfigMap < files[j].ConfigMap })
	return files, nil
}

func readServiceInfoFileFromConfigMap(cm corev1.ConfigMap, c *ServiceInfoFile) error {
	var fileName string
	for k, v := range cm.Annotations {
		if k == FileKey {
			fileName = v
		} else if k == PermissionsKey {
			c.Permissions = v
		} else if k == PathKey {
			c.Path = v
		}
	}
	if fileName == "" || c.Path == "" {
		return fmt.Errorf("serviceinfo file name and destination path are required: %s", cm.Name)
	}
	if _, ok := cm.BinaryData[fileName]; !ok {
		return fmt.Errorf("configmap '%s' does not contain file '%s'", cm.Name, fileName)
	}
	c.SourcePath = fmt.Sprintf(FilePathTemplate, cm.Name, fileName)
	c.ConfigMap = cm.Name
	return nil
}
