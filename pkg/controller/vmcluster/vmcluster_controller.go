package vmcluster

import (
	"context"
	"github.com/VictoriaMetrics/operator/conf"
	victoriametricsv1beta1 "github.com/VictoriaMetrics/operator/pkg/apis/victoriametrics/v1beta1"
	"github.com/VictoriaMetrics/operator/pkg/controller/factory"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

var log = logf.Log.WithName("controller_vmcluster")

// Add creates a new VMCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVmCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("vmcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	if err = c.Watch(&source.Kind{Type: &victoriametricsv1beta1.VMCluster{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}
	for _, s := range []runtime.Object{&v1.Deployment{}, &v1.StatefulSet{}, &corev1.Service{}} {
		if err = c.Watch(&source.Kind{Type: s}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &victoriametricsv1beta1.VMCluster{},
		}); err != nil {
			return err
		}
	}
	return nil
}

// blank assignment to verify that ReconcileVmCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVmCluster{}

// ReconcileVmCluster reconciles a VMCluster object
type ReconcileVmCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVmCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling VMCluster")
	ctx := context.TODO()
	cluster := &victoriametricsv1beta1.VMCluster{}
	if err := r.client.Get(ctx, request.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	//TODO move validate to admission webhook
	//if err := cluster.Validate(); err != nil {
	//	return reconcile.Result{}, err
	//}
	//first update storage - expand if needed and wait for ready status
	status, err := factory.CreateOrUpdateVMCluster(ctx, cluster, r.client, conf.MustGetBaseConfig())
	if err != nil {
		log.Error(err, "cannot update or create vmcluster")
		return reconcile.Result{}, err
	}
	if status == victoriametricsv1beta1.ClusterStatusExpanding {
		log.Info("cluster still expanding requeue request")
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: time.Second * 10,
		}, nil
	}

	//update select and wait for ready

	//update insert
	//ctl := &vmController{client: r.client, cluster: cluster}
	//if r, err := ctl.reconcileStorageLoop(ctx); err != nil || r != emptyReconcile {
	//	return r, err
	//}
	return reconcile.Result{}, nil
}
