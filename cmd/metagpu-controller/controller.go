package main

import (
	"context"
	v1 "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/cmd/metagpu-controller/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("metagpu")
	mg := v1.MetaGpu{}
	name := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}
	if err := r.Get(context.Background(), name, &mg); err != nil {
		l.Error(err, "error fetching metagpu")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func main() {
	ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	if err := v1.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&v1.MetaGpu{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})

	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
