/*
Copyright 2021.

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
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"

	"k8s.io/apimachinery/pkg/runtime"
	batchv1 "postgresql-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	deleteTable = "DROP TABLE IF EXISTS "
)

// PostgreSQLReconciler reconciles a PostgreSQL object
type PostgreSQLReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	DatabaseConnectionPool *pgxpool.Pool
}

//+kubebuilder:rbac:groups=batch.hub.docker.com,resources=postgresqls,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.hub.docker.com,resources=postgresqls/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.hub.docker.com,resources=postgresqls/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PostgreSQL object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PostgreSQLReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("PostgreSQL", req.NamespacedName)
	var pSpec batchv1.PostgreSQL
	var err error
	if err = r.Get(ctx, req.NamespacedName, &pSpec); err != nil {
		log.Error(err, "unable to fetch operator")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	tablename := pSpec.ObjectMeta.GetName()
	_, err = r.DatabaseConnectionPool.Exec(context.Background(), deleteTable + tablename )
	if err != nil {
		err = fmt.Errorf("failed to delete table during reinitialization because: %w", err)
		log.Error(err, "deletion failed")
		return ctrl.Result{}, err
	}
	for colName, colVal := range pSpec.Spec {

	}


	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.PostgreSQL{}).
		Complete(r)
}
