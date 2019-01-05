package vpc

import (
	"fmt"
	"github.com/kfirz/gitzup/internal/reconciler"
	"github.com/kfirz/gitzup/internal/util/gcp"
	"github.com/kfirz/gitzup/pkg/apis/gcp/v1beta1"
	"github.com/pkg/errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add creates a new IpAddress Controller and adds it to the Manager with default RBAC. The Manager will set fields on
// the Controller and Start it when the Manager is Started.
//
// +kubebuilder:rbac:groups=gcp.gitzup.com,resources=ipaddresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gcp.gitzup.com,resources=ipaddresses/status,verbs=get;list;watch;create;update;patch;delete
func AddIpAddress(mgr manager.Manager) error {

	r := reconciler.New("IpAddresses", mgr, &ResourceAdapter{})
	err := r.Start()
	if err != nil {
		return errors.Wrapf(err, "failed starting IpAddress reconciler")
	}

	return nil
}

// Adapter for working with GCP reserved IP addresses.
type ResourceAdapter struct {
	r *reconciler.Reconciler
}

// Ensure our resource adapter struct implements the reconcile.ResourceAdapter interface
var _ reconciler.ObjectAdapter = &ResourceAdapter{}

func (a *ResourceAdapter) IsCleanupOnDeletion() bool {
	return true
}

func (a *ResourceAdapter) Inject(r *reconciler.Reconciler) {
	a.r = r
}

func (a *ResourceAdapter) CreateObject() interface{} {
	return &v1beta1.IpAddress{}
}

func (a *ResourceAdapter) CreateList() interface{} {
	return &v1beta1.IpAddressList{}
}

func (a *ResourceAdapter) GetListItems(list interface{}) ([]interface{}, error) {
	if objList, ok := list.(*v1beta1.IpAddressList); ok {
		items := make([]interface{}, len(objList.Items))
		for i := 0; i < len(objList.Items); i++ {
			items[i] = &objList.Items[i]
		}
		return items, nil
	}
	return nil, errors.Errorf("received '%s' list, instead of '*IpAddressList'", reflect.TypeOf(list))
}

func (a *ResourceAdapter) GetObjectMeta(obj interface{}) *metav1.ObjectMeta {
	if o, ok := obj.(*v1beta1.IpAddress); ok {
		return &o.ObjectMeta
	}
	panic(errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj)))
}

func (a *ResourceAdapter) GetRuntimeObject(obj interface{}) runtime.Object {
	if o, ok := obj.(*v1beta1.IpAddress); ok {
		return o
	} else if o, ok := obj.(*v1beta1.IpAddressList); ok {
		return o
	}
	panic(errors.Errorf("received '%s' object/list instead of '*IpAddress' or '*IpAddressList'", reflect.TypeOf(obj)))
}

func (a *ResourceAdapter) CreateResource(obj interface{}) (interface{}, error) {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	addressName := o.ObjectMeta.Namespace + "-" + o.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateComputeClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP compute client")
	}

	// Describe the address
	addr := compute.Address{
		Name:        addressName,
		Description: fmt.Sprintf("IP address for Kubernetes object '%s/%s'", o.ObjectMeta.Namespace, o.ObjectMeta.Name),
		NetworkTier: o.Spec.NetworkTier,
	}

	// Create it, and get a long-running operation to wait for
	var op *compute.Operation
	if o.Spec.Region == "" {
		addr.IpVersion = o.Spec.IpVersion
		op, err = svc.GlobalAddresses.Insert(o.Spec.ProjectId, &addr).Do()
	} else {
		op, err = svc.Addresses.Insert(o.Spec.ProjectId, o.Spec.Region, &addr).Do()
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating resource")
	}

	// Wait until operation is done or fails
	err = gcp.WaitForComputeOperation(o.Spec.ProjectId, op)
	if err != nil {
		return nil, errors.Wrapf(err, "resource creation failed")
	}

	return a.RetrieveResource(obj)
}

func (a *ResourceAdapter) RetrieveResource(obj interface{}) (interface{}, error) {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	addressName := o.ObjectMeta.Namespace + "-" + o.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateComputeClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP compute client")
	}

	// Global/Regional reserved address?
	var r interface{}
	if o.Spec.Region == "" {
		r, err = svc.GlobalAddresses.Get(o.Spec.ProjectId, addressName).Do()
	} else {
		r, err = svc.Addresses.Get(o.Spec.ProjectId, o.Spec.Region, addressName).Do()
	}

	// If an error occurred, check the type - it might be 404 in which case we need to return nil
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed fetching resource")
	}

	return r, nil
}

func (a *ResourceAdapter) UpdateResource(obj interface{}, resource interface{}) (interface{}, error) {
	if addr, ok := resource.(*compute.Address); !ok {
		return nil, errors.Errorf("received '%s' resource instead of '*compute.Address'", reflect.TypeOf(resource))
	} else {
		err := a.DeleteResource(obj)
		if err != nil {
			return nil, errors.Wrapf(err, "failed deleting GCP reserved IP address '%s'", addr.Name)
		}

		addr, err := a.CreateResource(obj)
		if err != nil {
			return nil, errors.Wrapf(err, "failed creating GCP reserved IP address")
		}

		return addr, nil
	}
}

func (a *ResourceAdapter) DeleteResource(obj interface{}) error {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	addressName := o.ObjectMeta.Namespace + "-" + o.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateComputeClient()
	if err != nil {
		return errors.Wrapf(err, "failed creating GCP compute client")
	}

	// Delete it, and get a long-running operation to wait for
	var op *compute.Operation
	if o.Spec.Region == "" {
		op, err = svc.GlobalAddresses.Delete(o.Spec.ProjectId, addressName).Do()
		if err != nil {
			return errors.Wrapf(err, "failed deleting global IP address")
		}
	} else {
		op, err = svc.Addresses.Delete(o.Spec.ProjectId, o.Spec.Region, addressName).Do()
		if err != nil {
			return errors.Wrapf(err, "failed deleting regional IP address")
		}
	}

	// Wait until operation is done or fails
	err = gcp.WaitForComputeOperation(o.Spec.ProjectId, op)
	if err != nil {
		return errors.Wrapf(err, "operation failed")
	}

	return nil
}

func (a *ResourceAdapter) IsUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return false, errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	if addr, ok := resource.(*compute.Address); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*compute.Address'", reflect.TypeOf(resource))
	} else {
		fullRegion := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s", o.Spec.ProjectId, o.Spec.Region)
		if addr.NetworkTier != o.Spec.NetworkTier {
			return true, nil
		} else if addr.Region == "" && addr.IpVersion != o.Spec.IpVersion {
			return true, nil
		} else if addr.Region != fullRegion {
			return true, nil
		}
		return false, nil
	}
}

func (a *ResourceAdapter) IsStatusUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return false, errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	if addr, ok := resource.(*compute.Address); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*compute.Address'", reflect.TypeOf(resource))
	} else if addr.Address != o.Status.Address {
		return true, nil
	}
	return false, nil
}

func (a *ResourceAdapter) UpdateObjectStatus(obj interface{}, resource interface{}) error {
	o, ok := obj.(*v1beta1.IpAddress)
	if !ok {
		return errors.Errorf("received '%s' object instead of '*IpAddress'", reflect.TypeOf(obj))
	}

	if addr, ok := resource.(*compute.Address); !ok {
		return errors.Errorf("received '%s' resource instead of '*compute.Address'", reflect.TypeOf(resource))
	} else {
		o.Status.Address = addr.Address
		return nil
	}
}
