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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	fdov1 "github.com/empovit/fdo-operators/api/v1"
)

// FDOManufacturingReconciler reconciles a FDOManufacturing object
type FDOManufacturingReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=fdo.example.com,resources=fdomanufacturings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=fdo.example.com,resources=fdomanufacturings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=fdo.example.com,resources=fdomanufacturings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FDOManufacturing object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *FDOManufacturingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("")
	log = logf.Log.WithName("fdomanufacturing_controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconciling FDO manufacturing server")

	fdoServer := &fdov1.FDOManufacturing{}
	err := r.Get(ctx, req.NamespacedName, fdoServer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("FDOManufacturing resource not found. Ignoring since object must have been deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get FDOManufacturing resource")
		return ctrl.Result{}, err
	}

	fdoServer = r.setDefaultValues(fdoServer)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FDOManufacturingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fdov1.FDOManufacturing{}).
		Complete(r)
}

func (r *FDOManufacturingReconciler) generateConfig(fdoServer *fdov1.FDOManufacturing) string {
	return fmt.Sprintf(`
		%s`,
		"webAppWarFileName,webAppSourceRepositoryURL,webAppSourceRepositoryRef,webAppSourceRepositoryContextDir",
	)
}

func (r *FDOManufacturingReconciler) setDefaultValues(fdoServer *fdov1.FDOManufacturing) *fdov1.FDOManufacturing {
	return nil
}
