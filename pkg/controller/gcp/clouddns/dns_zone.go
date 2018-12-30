package clouddns

import (
	"fmt"
	"github.com/kfirz/gitzup/internal/reconciler"
	"github.com/kfirz/gitzup/internal/util/gcp"
	"github.com/kfirz/gitzup/pkg/apis/gcp/v1beta1"
	"github.com/pkg/errors"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add creates a new DnsZone Controller and adds it to the Manager with default RBAC. The Manager will set fields on
// the Controller and Start it when the Manager is Started.
//
// +kubebuilder:rbac:groups=gcp.gitzup.com,resources=dnszones,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gcp.gitzup.com,resources=dnszones/status,verbs=get;list;watch;create;update;patch;delete
func AddDnsZone(mgr manager.Manager) error {

	r := reconciler.New("DnsZone", mgr, &ResourceAdapter{})
	err := r.Start()
	if err != nil {
		return errors.Wrapf(err, "failed starting DnsZone reconciler")
	}

	return nil
}

// External resource adapter
type ResourceAdapter struct {
}

// Ensure our resource adapter struct implements the reconcile.ResourceAdapter interface
var _ reconciler.ObjectAdapter = &ResourceAdapter{}

func (a *ResourceAdapter) CreateObject() interface{} {
	return &v1beta1.DnsZone{}
}

func (a *ResourceAdapter) CreateList() interface{} {
	return &v1beta1.DnsZoneList{}
}

func (a *ResourceAdapter) GetListItems(list interface{}) ([]interface{}, error) {
	if objList, ok := list.(*v1beta1.DnsZoneList); ok {
		items := make([]interface{}, len(objList.Items))
		for i := 0; i < len(objList.Items); i++ {
			items[i] = &objList.Items[i]
		}
		return items, nil
	}
	return nil, errors.Errorf("received '%s' list, instead of '*DnsZoneList'", reflect.TypeOf(list))
}

func (a *ResourceAdapter) GetObjectMeta(obj interface{}) *metav1.ObjectMeta {
	if o, ok := obj.(*v1beta1.DnsZone); ok {
		return &o.ObjectMeta
	}
	panic(errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj)))
}

func (a *ResourceAdapter) GetRuntimeObject(obj interface{}) runtime.Object {
	if o, ok := obj.(*v1beta1.DnsZone); ok {
		return o
	} else if o, ok := obj.(*v1beta1.DnsZoneList); ok {
		return o
	}
	panic(errors.Errorf("received '%s' object/list instead of '*DnsZone' or '*DnsZoneList'", reflect.TypeOf(obj)))
}

func (a *ResourceAdapter) CreateResource(obj interface{}) (interface{}, error) {
	o, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Describe the resource
	zone := dns.ManagedZone{
		Name:        o.ObjectMeta.Namespace + "-" + o.ObjectMeta.Name,
		Description: fmt.Sprintf("%s/%s", o.ObjectMeta.Namespace, o.ObjectMeta.Name),
		DnsName:     o.Spec.DnsName,
	}

	// Create it
	_, err = svc.ManagedZones.Create(o.Spec.ProjectId, &zone).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "resource creation failed")
	}

	return a.RetrieveResource(obj)
}

func (a *ResourceAdapter) RetrieveResource(obj interface{}) (interface{}, error) {
	o, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Global/Regional reserved address?
	managedZone, err := svc.ManagedZones.Get(o.Spec.ProjectId, o.ObjectMeta.Namespace+"-"+o.ObjectMeta.Name).Do()
	if err != nil {
		// If an error occurred, check the type - it might be 404 in which case we need to return nil
		if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed fetching resource")
	}

	return managedZone, nil
}

func (a *ResourceAdapter) UpdateResource(obj interface{}, resource interface{}) (interface{}, error) {
	var ok bool

	var o *v1beta1.DnsZone
	if o, ok = obj.(*v1beta1.DnsZone); !ok {
		return nil, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}
	if _, ok = obj.(*dns.ManagedZone); !ok {
		return nil, errors.Errorf("received '%s' resource instead of '*dns.ManagedZone'", reflect.TypeOf(resource))
	}

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Describe the patch
	zone := dns.ManagedZone{
		Name:        o.ObjectMeta.Namespace + "-" + o.ObjectMeta.Name,
		Description: fmt.Sprintf("%s/%s", o.ObjectMeta.Namespace, o.ObjectMeta.Name),
		DnsName:     o.Spec.DnsName,
	}

	// Patch the zone
	op, err := svc.ManagedZones.Patch(o.Spec.ProjectId, o.ObjectMeta.Namespace+"-"+o.ObjectMeta.Name, &zone).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed updating GCP zone '%s'", zone.Name)
	}

	// Wait for the operation to complete successfully
	err = gcp.WaitForDnsOperation(o.Spec.ProjectId, op)
	if err != nil {
		return nil, errors.Wrapf(err, "operation failed")
	}

	return a.RetrieveResource(obj)
}

func (a *ResourceAdapter) DeleteResource(obj interface{}) error {
	o, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return errors.Wrapf(err, "failed creating GCP DNS client")
	}

	// Delete it
	err = svc.ManagedZones.Delete(o.Spec.ProjectId, o.ObjectMeta.Namespace+"-"+o.ObjectMeta.Name).Do()
	if err != nil {
		return errors.Wrapf(err, "failed deleting DNS zone")
	}

	return nil
}

func (a *ResourceAdapter) IsUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	var ok bool

	var o *v1beta1.DnsZone
	if o, ok = obj.(*v1beta1.DnsZone); !ok {
		return false, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var r *dns.ManagedZone
	if r, ok = resource.(*dns.ManagedZone); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*dns.ManagedZone'", reflect.TypeOf(resource))
	}

	// Compare
	if r.DnsName != o.Spec.DnsName {
		return true, nil
	}
	if r.Description != fmt.Sprintf("%s/%s", o.ObjectMeta.Namespace, o.ObjectMeta.Name) {
		return true, nil
	}

	return false, nil
}

func (a *ResourceAdapter) IsStatusUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	var ok bool

	var o *v1beta1.DnsZone
	if o, ok = obj.(*v1beta1.DnsZone); !ok {
		return false, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var r *dns.ManagedZone
	if r, ok = resource.(*dns.ManagedZone); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*dns.ManagedZone'", reflect.TypeOf(resource))
	}

	// Compare
	if r.Id != o.Status.Id {
		return true, nil
	}

	return false, nil
}

func (a *ResourceAdapter) UpdateObjectStatus(obj interface{}, resource interface{}) error {
	var ok bool

	var o *v1beta1.DnsZone
	if o, ok = obj.(*v1beta1.DnsZone); !ok {
		return errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var r *dns.ManagedZone
	if r, ok = resource.(*dns.ManagedZone); !ok {
		return errors.Errorf("received '%s' resource instead of '*dns.ManagedZone'", reflect.TypeOf(resource))
	}

	o.Status.Id = r.Id
	return nil
}
