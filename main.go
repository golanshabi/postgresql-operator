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

package main

import (
	"context"
	"flag"
	"os"
	"postgresql-operator/controllers"

	batchv1 "postgresql-operator/api/v1"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var scheme = runtime.NewScheme()

const (
	environmentVariableDatabaseURL = "DATABASE_URL"
	portNum                        = 9443
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
	utilruntime.Must(batchv1.AddToScheme(scheme))
}

func doMain() int {
	log := ctrl.Log.WithName("cmd")

	var metricsAddr string

	var enableLeaderElection bool

	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   portNum,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "a790ce0a.hub.docker.com",
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	dbConnectionPool, i, done := connectToDB(log)

	defer dbConnectionPool.Close()

	if done {
		return i
	}

	if err = (&controllers.PostgreSQLReconciler{
		Client:                 mgr.GetClient(),
		Log:                    ctrl.Log.WithName("controllers").WithName("PostgreSQL"),
		Scheme:                 mgr.GetScheme(),
		DatabaseConnectionPool: dbConnectionPool,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "PostgreSQL")

		return 1
	}
	//+kubebuilder:scaffold:builder

	i2, done2 := startManager(mgr, log)
	if done2 {
		return i2
	}

	return 0
}

func startManager(mgr manager.Manager, log logr.Logger) (int, bool) {
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")

		return 1, true
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")

		return 1, true
	}

	log.Info("starting manager")

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")

		return 1, true
	}

	return 0, false
}

func connectToDB(log logr.Logger) (*pgxpool.Pool, int, bool) {
	databaseURL, found := os.LookupEnv(environmentVariableDatabaseURL)
	if !found {
		log.Error(nil, "Not found:", "environment variable", environmentVariableDatabaseURL)

		return nil, 1, true
	}

	dbConnectionPool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Error(err, "Failed to connect to the database")

		return nil, 1, true
	}

	return dbConnectionPool, 0, false
}

func main() {
	os.Exit(doMain())
}
