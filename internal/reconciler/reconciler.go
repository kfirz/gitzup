package reconciler

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kfirz/gitzup/internal/util"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"time"
)

const (
	finalizerName = "finalizers.gitzup.com"
)

var (
	externalWatchInterval = 30 * time.Second
)

// Adapts a Kubernetes object to an external resource (eg. a GCP or AWS machine, IP address, etc)
type ObjectAdapter interface {
	Inject(reconciler *Reconciler)
	IsCleanupOnDeletion() bool
	FetchObject(ctx context.Context, request reconcile.Request) (interface{}, *metav1.ObjectMeta, runtime.Object, error)
	CreateObject() interface{}
	CreateList() interface{}
	GetListItems(list interface{}) ([]interface{}, error)
	GetObjectMeta(obj interface{}) *metav1.ObjectMeta
	GetRuntimeObject(obj interface{}) runtime.Object
	CreateResource(obj interface{}) (interface{}, error)
	RetrieveResource(obj interface{}) (interface{}, error)
	UpdateResource(obj interface{}, resource interface{}) (interface{}, error)
	DeleteResource(obj interface{}) error
	IsUpdateNeeded(obj interface{}, resource interface{}) (bool, error)
	IsStatusUpdateNeeded(obj interface{}, resource interface{}) (bool, error)
	UpdateObjectStatus(obj interface{}, resource interface{}) error
}

// Reconciles differences between an admitted Kubernetes object and its physical representation in the cloud provider.
type Reconciler struct {
	Name string
	manager.Manager
	client.Client
	record.EventRecorder
	ObjectAdapter
	Scheme                     *runtime.Scheme
	Log                        logr.Logger
	Debug                      logr.InfoLogger
	externalReconcileChan      chan event.GenericEvent
	closeExternalReconcileChan chan struct{}
}

// Ensure our Reconciler struct implements the reconcile.Reconciler interface
var _ reconcile.Reconciler = &Reconciler{}

func New(name string, mgr manager.Manager, adapter ObjectAdapter) *Reconciler {
	r := &Reconciler{
		Name:                       name,
		Manager:                    mgr,
		Client:                     mgr.GetClient(),
		EventRecorder:              mgr.GetRecorder("gitzup"),
		ObjectAdapter:              adapter,
		Scheme:                     mgr.GetScheme(),
		Log:                        log.Log.WithName("reconciler").WithName(name),
		Debug:                      log.Log.WithName("reconciler").WithName(name).V(1),
		externalReconcileChan:      make(chan event.GenericEvent),
		closeExternalReconcileChan: make(chan struct{}),
	}
	adapter.Inject(r)
	return r
}

func (r *Reconciler) WarnEvent(obj runtime.Object, reason string, msg string, msgArgs ...interface{}) {
	r.EventRecorder.Eventf(
		obj,
		v1.EventTypeWarning,
		reason,
		msg,
		msgArgs...,
	)
}

func (r *Reconciler) InfoEvent(obj runtime.Object, reason string, msg string, msgArgs ...interface{}) {
	r.EventRecorder.Eventf(
		obj,
		v1.EventTypeNormal,
		reason,
		msg,
		msgArgs...,
	)
}

func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()
	mgrClient := r.Manager.GetClient()
	adapter := r.ObjectAdapter

	// Fetch the object
	object, objectMeta, runtimeObject, err := adapter.FetchObject(ctx, request)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Object not found, return (cleanup logic is implemented through finalizers)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, errors.Wrapf(err, "could not find object")
	}

	// Is the object being deleted?
	if !objectMeta.GetDeletionTimestamp().IsZero() {

		// if our finalizer hasn't been executed yet, run it now and remove it from the finalizers list
		if adapter.IsCleanupOnDeletion() && util.ContainsString(objectMeta.GetFinalizers(), finalizerName) {

			// Discover whether the resource exists
			res, err := adapter.RetrieveResource(object)
			if err != nil {
				r.WarnEvent(runtimeObject, "ResourceRetrievalError", "Could not check if resource exists (for deletion)")
				return reconcile.Result{}, errors.Wrapf(err, "could not retrieve external resource")
			}

			// If it exists, delete it
			if res != nil {
				if err := adapter.DeleteResource(object); err != nil {
					r.WarnEvent(runtimeObject, "ResourceDeletionError", "Could not delete resource")
					return reconcile.Result{}, errors.Wrapf(err, "could not delete external resource")
				}
				r.InfoEvent(runtimeObject, "ExternalResourceDeleted", "Deleted external resource")
			}
		}

		// remove our finalizer from the list and update it.
		objectMeta.SetFinalizers(util.RemoveString(objectMeta.GetFinalizers(), finalizerName))
		if err := mgrClient.Update(ctx, runtimeObject); err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not remove object finalizer")
			return reconcile.Result{}, errors.Wrapf(err, "could not remove finalizer")
		}

		// All done
		return reconcile.Result{}, nil
	}

	// Make sure our finalizer is listed in the object's finalizers list
	if !util.ContainsString(objectMeta.GetFinalizers(), finalizerName) {
		objectMeta.SetFinalizers(append(objectMeta.GetFinalizers(), finalizerName))
		if err := mgrClient.Update(ctx, runtimeObject); err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not add object finalizer")
			return reconcile.Result{}, errors.Wrapf(err, "could not add finalizer")
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Fetch resource
	res, err := adapter.RetrieveResource(object)
	if err != nil {
		r.WarnEvent(runtimeObject, "ResourceRetrievalError", "Could not retrieve resource")
		return reconcile.Result{}, errors.Wrapf(err, "could not retrieve external resource")
	}

	// If resource is missing, create it
	if res == nil {

		// Create resource
		res, err := adapter.CreateResource(object)
		if err != nil {
			r.WarnEvent(runtimeObject, "ResourceCreationError", "Could not create resource")
			return reconcile.Result{}, errors.Wrapf(err, "could not create external resource")
		}
		r.InfoEvent(runtimeObject, "ExternalResourceCreated", "Created external resource")

		// Updating the status
		err = adapter.UpdateObjectStatus(object, res)
		if err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not update object status")
			return reconcile.Result{}, errors.Wrapf(err, "could not update object status")
		}

		// Save the status update
		err = r.Manager.GetClient().Status().Update(ctx, runtimeObject)
		if err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not persist object status update")
			return reconcile.Result{}, errors.Wrapf(err, "could not persist object status")
		}

		return reconcile.Result{}, nil
	}

	// Check if an update to the external resource is needed
	stale, err := adapter.IsUpdateNeeded(object, res)
	if err != nil {
		r.WarnEvent(runtimeObject, "StalenessCheckError", "Could not check if resource needs to be updated")
		return reconcile.Result{}, errors.Wrapf(err, "could not check if resource or object are stale")
	}

	// If the external resource is stale
	if stale {
		_, err := adapter.UpdateResource(object, res)
		if err != nil {
			r.WarnEvent(runtimeObject, "ResourceUpdateError", "Could not update resource")
			return reconcile.Result{}, errors.Wrapf(err, "could not update external resource")
		}
		r.InfoEvent(runtimeObject, "ExternalResourceUpdated", "Updated external resource")
	}

	// Check if an update to the object status is needed
	statusStale, err := adapter.IsStatusUpdateNeeded(object, res)
	if err != nil {
		r.WarnEvent(runtimeObject, "StalenessCheckError", "Could not check if object status needs to be updated")
		return reconcile.Result{}, errors.Wrapf(err, "could not check if object status is stale")
	}

	// If the object status is stale
	if statusStale {

		// Updating the status
		err = adapter.UpdateObjectStatus(object, res)
		if err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not update object status")
			return reconcile.Result{}, errors.Wrapf(err, "could not update object status")
		}

		// Save the status update
		err = r.Manager.GetClient().Status().Update(ctx, runtimeObject)
		if err != nil {
			r.WarnEvent(runtimeObject, "ObjectUpdateError", "Could not persist object status update")
			return reconcile.Result{}, errors.Wrapf(err, "could not persist object status")
		}

	}

	// all ok
	return reconcile.Result{}, nil
}

func (r *Reconciler) Start() error {
	// Create a new controller
	c, err := controller.New(r.Name+"-controller", r.Manager, controller.Options{Reconciler: r})
	if err != nil {
		return errors.Wrapf(err, "failed creating controller for '%s' reconciler", r.Name)
	}

	// Watch for changes to matching Kubernetes objects
	object := r.ObjectAdapter.CreateObject()
	runtimeObject := r.ObjectAdapter.GetRuntimeObject(object)
	err = c.Watch(&source.Kind{Type: runtimeObject}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrapf(err, "failed watching Kubernetes objects for '%s' reconciler", r.Name)
	}

	// Watch for reconciliation requests caused from external resource changes
	err = c.Watch(&source.Channel{Source: r.externalReconcileChan}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return errors.Wrapf(err, "failed watching for events due to external resource changes for '%s' reconciler", r.Name)
	}

	// Start a thread to periodically check the external resource
	go wait.Until(
		func() {
			ctx := context.TODO()
			list := r.ObjectAdapter.CreateList()
			runtimeList := r.ObjectAdapter.GetRuntimeObject(list)

			// List the controller objects
			err := r.Client.List(ctx, nil, runtimeList)
			if err != nil {
				r.Log.Error(err, "Failed listing objects")
				return
			}

			// Extract the items from the list
			objects, err := r.ObjectAdapter.GetListItems(list)
			if err != nil {
				r.Log.Error(err, "Failed extracting items from list")
				return
			}

			// Iterate over the items, and check whether reconciliation is required for any of them
			for _, object := range objects {
				objectMeta := r.ObjectAdapter.GetObjectMeta(object)
				runtimeObject := r.ObjectAdapter.GetRuntimeObject(object)

				// Ignore objects being deleted
				if !objectMeta.GetDeletionTimestamp().IsZero() {
					r.Debug.Info("Ignoring object being deleted", "object", object)
					continue
				}

				// Fetch resource
				res, err := r.ObjectAdapter.RetrieveResource(object)
				if err != nil {
					r.WarnEvent(runtimeObject, "ResourceRetrievalError", "Could not retrieve resource")
					r.Log.Error(err, "Failed retrieving resource", "object", object)
					continue
				}

				// If resource is missing, update the Kubernetes object so it will be reconciled
				if res == nil {
					r.externalReconcileChan <- event.GenericEvent{Meta: objectMeta, Object: runtimeObject}
					continue
				}

				// If resource exists, check if reconciliation is required
				resourceStale, err := r.ObjectAdapter.IsUpdateNeeded(object, res)
				if err != nil {
					r.WarnEvent(runtimeObject, "StalenessCheckError", "Could not check if resource needs to be updated")
					r.Log.Error(err, "Failed checking object & resource staleness", "object", object)
					continue
				}

				// If reconciliation is needed, send a request
				if resourceStale {
					r.externalReconcileChan <- event.GenericEvent{Meta: objectMeta, Object: runtimeObject}
				}

				// Check if an update to the object status is needed
				statusStale, err := r.ObjectAdapter.IsStatusUpdateNeeded(object, res)
				if err != nil {
					r.WarnEvent(runtimeObject, "StalenessCheckError", "Could not check if object status needs to be updated")
					r.Log.Error(err, "Failed checking status staleness", "object", object)
					continue
				}

				// If the object status is stale
				if statusStale {
					r.externalReconcileChan <- event.GenericEvent{Meta: objectMeta, Object: runtimeObject}
				}
			}
		},
		externalWatchInterval,
		r.closeExternalReconcileChan,
	)

	return nil
}
