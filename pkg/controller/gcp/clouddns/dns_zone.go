package clouddns

import (
	"context"
	"fmt"
	"github.com/kfirz/gitzup/internal/reconciler"
	"github.com/kfirz/gitzup/internal/util"
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
	r *reconciler.Reconciler
}

type zoneAndRecords struct {
	Zone    *dns.ManagedZone
	Records []*dns.ResourceRecordSet
}

// Ensure our resource adapter struct implements the reconcile.ResourceAdapter interface
var _ reconciler.ObjectAdapter = &ResourceAdapter{}

func (a *ResourceAdapter) Inject(r *reconciler.Reconciler) {
	a.r = r
}

func (a *ResourceAdapter) IsCleanupOnDeletion() bool {
	return false
}

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
	kzone, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}
	zoneName := kzone.ObjectMeta.Namespace + "-" + kzone.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Describe the resource
	zone := dns.ManagedZone{
		Name:        zoneName,
		Description: fmt.Sprintf("%s/%s", kzone.ObjectMeta.Namespace, kzone.ObjectMeta.Name),
		DnsName:     kzone.Spec.DnsName,
	}

	// Create it
	_, err = svc.ManagedZones.Create(kzone.Spec.ProjectId, &zone).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "resource creation failed")
	}

	// Create records
	change := &dns.Change{}
	for _, rec := range kzone.Spec.Records {
		change.Additions = append(change.Additions, &dns.ResourceRecordSet{
			Name:    rec.DnsName,
			Type:    rec.Type,
			Ttl:     rec.Ttl,
			Rrdatas: rec.Rrdatas,
		})
	}
	_, err = svc.Changes.Create(kzone.Spec.ProjectId, zoneName, change).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating DNS records")
	}

	return a.RetrieveResource(obj)
}

func (a *ResourceAdapter) RetrieveResource(obj interface{}) (interface{}, error) {
	kzone, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return nil, errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}
	zoneName := kzone.ObjectMeta.Namespace + "-" + kzone.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Retrieve the zone
	managedZone, err := svc.ManagedZones.Get(kzone.Spec.ProjectId, zoneName).Do()
	if err != nil {
		// If an error occurred, check the type - it might be 404 in which case we need to return nil
		if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed fetching resource")
	}

	// Retrieve the records
	recordSets := make([]*dns.ResourceRecordSet, 0)
	err = svc.ResourceRecordSets.List(kzone.Spec.ProjectId, zoneName).Pages(
		context.TODO(),
		func(resp *dns.ResourceRecordSetsListResponse) error {
			for _, recordSet := range resp.Rrsets {
				recordSets = append(recordSets, recordSet)
			}
			return nil
		})
	if err != nil {
		return nil, errors.Wrapf(err, "failed fetching zone records")
	}

	return &zoneAndRecords{Zone: managedZone, Records: recordSets}, nil
}

func (a *ResourceAdapter) UpdateResource(obj interface{}, resource interface{}) (interface{}, error) {
	var ok bool

	var kzone *v1beta1.DnsZone
	if kzone, ok = obj.(*v1beta1.DnsZone); !ok {
		return nil, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var rzone *zoneAndRecords
	if rzone, ok = resource.(*zoneAndRecords); !ok {
		return nil, errors.Errorf("received '%s' resource instead of '*zoneAndRecords'", reflect.TypeOf(resource))
	}

	zoneName := kzone.ObjectMeta.Namespace + "-" + kzone.ObjectMeta.Name

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP client")
	}

	// Describe the patch
	zone := dns.ManagedZone{
		Name:        zoneName,
		Description: fmt.Sprintf("%s/%s", kzone.ObjectMeta.Namespace, kzone.ObjectMeta.Name),
		DnsName:     kzone.Spec.DnsName,
	}

	// Patch the zone
	op, err := svc.ManagedZones.Patch(kzone.Spec.ProjectId, zoneName, &zone).Do()
	if err != nil {
		return nil, errors.Wrapf(err, "failed updating GCP zone '%s'", zone.Name)
	}

	// Wait for the operation to complete successfully
	err = gcp.WaitForDnsOperation(kzone.Spec.ProjectId, op)
	if err != nil {
		return nil, errors.Wrapf(err, "operation failed")
	}

	// Map kzone records by name
	krecordsByName := make(map[string]*v1beta1.DnsRecord)
	for _, krec := range kzone.Spec.Records {
		krecordsByName[krec.DnsName] = &v1beta1.DnsRecord{
			DnsName: krec.DnsName,
			Type:    krec.Type,
			Ttl:     krec.Ttl,
			Rrdatas: krec.Rrdatas,
		}
	}

	// Map rzone records by name
	rrecordsByName := make(map[string]*dns.ResourceRecordSet)
	for _, rrec := range rzone.Records {
		rrecordsByName[rrec.Name] = &dns.ResourceRecordSet{
			Kind:             rrec.Kind,
			Name:             rrec.Name,
			Type:             rrec.Type,
			Ttl:              rrec.Ttl,
			Rrdatas:          rrec.Rrdatas,
			ForceSendFields:  rrec.ForceSendFields,
			NullFields:       rrec.NullFields,
			SignatureRrdatas: rrec.SignatureRrdatas,
		}
	}

	// Compare spec DNS records against the actual zone's DNS records, and update accordingly
	for krecName, krec := range krecordsByName {
		if rrec, ok := rrecordsByName[krecName]; ok {
			if krec.Type != rrec.Type || krec.Ttl != rrec.Ttl || !util.StringSlicesEqual(krec.Rrdatas, rrec.Rrdatas) {
				changeSet := &dns.Change{
					Deletions: []*dns.ResourceRecordSet{rrec},
					Additions: []*dns.ResourceRecordSet{
						{
							Name:    rrec.Name,
							Type:    krec.Type,
							Ttl:     krec.Ttl,
							Rrdatas: krec.Rrdatas,
						},
					},
				}
				_, err = svc.Changes.Create(kzone.Spec.ProjectId, zoneName, changeSet).Do()
				if err != nil {
					return nil, errors.Wrapf(err, "failed updating DNS record")
				}
			}
		} else {
			changeSet := &dns.Change{
				Additions: []*dns.ResourceRecordSet{{
					Name:    krec.DnsName,
					Type:    krec.Type,
					Ttl:     krec.Ttl,
					Rrdatas: krec.Rrdatas,
				}},
			}
			_, err = svc.Changes.Create(kzone.Spec.ProjectId, zoneName, changeSet).Do()
			if err != nil {
				return nil, errors.Wrapf(err, "failed creating DNS record")
			}
		}
	}

	return a.RetrieveResource(obj)
}

func (a *ResourceAdapter) DeleteResource(obj interface{}) error {
	kzone, ok := obj.(*v1beta1.DnsZone)
	if !ok {
		return errors.Errorf("received '%s' object instead of '*DnsZone'", reflect.TypeOf(obj))
	}

	// Create Google APIs client
	svc, err := gcp.CreateDnsClient()
	if err != nil {
		return errors.Wrapf(err, "failed creating GCP DNS client")
	}

	// Delete it
	err = svc.ManagedZones.Delete(kzone.Spec.ProjectId, kzone.ObjectMeta.Namespace+"-"+kzone.ObjectMeta.Name).Do()
	if err != nil {
		return errors.Wrapf(err, "failed deleting DNS zone")
	}

	return nil
}

func (a *ResourceAdapter) IsUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	var ok bool

	var kzone *v1beta1.DnsZone
	if kzone, ok = obj.(*v1beta1.DnsZone); !ok {
		return false, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var rzone *zoneAndRecords
	if rzone, ok = resource.(*zoneAndRecords); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*zoneAndRecords'", reflect.TypeOf(resource))
	}

	// Compare zone
	if rzone.Zone.DnsName != kzone.Spec.DnsName {
		return true, nil
	}
	if rzone.Zone.Description != fmt.Sprintf("%s/%s", kzone.ObjectMeta.Namespace, kzone.ObjectMeta.Name) {
		return true, nil
	}

	// Map kzone records by name
	krecordsByName := make(map[string]*v1beta1.DnsRecord)
	for _, krec := range kzone.Spec.Records {
		krecordsByName[krec.DnsName] = &v1beta1.DnsRecord{
			DnsName: krec.DnsName,
			Type:    krec.Type,
			Ttl:     krec.Ttl,
			Rrdatas: krec.Rrdatas,
		}
	}

	// Map rzone records by name
	rrecordsByName := make(map[string]*dns.ResourceRecordSet)
	for _, rrec := range rzone.Records {
		rrecordsByName[rrec.Name] = rrec
		rrecordsByName[rrec.Name] = &dns.ResourceRecordSet{
			Kind:             rrec.Kind,
			Name:             rrec.Name,
			Type:             rrec.Type,
			Ttl:              rrec.Ttl,
			Rrdatas:          rrec.Rrdatas,
			ForceSendFields:  rrec.ForceSendFields,
			NullFields:       rrec.NullFields,
			SignatureRrdatas: rrec.SignatureRrdatas,
		}
	}

	// Compare spec DNS records against the actual zone's DNS records
	for krecName, krec := range krecordsByName {
		if rrec, ok := rrecordsByName[krecName]; ok {
			if krec.Type != rrec.Type || krec.Ttl != rrec.Ttl || !util.StringSlicesEqual(krec.Rrdatas, rrec.Rrdatas) {
				return true, nil
			}
		} else {
			return true, nil
		}
	}

	return false, nil
}

func (a *ResourceAdapter) IsStatusUpdateNeeded(obj interface{}, resource interface{}) (bool, error) {
	var ok bool

	var kzone *v1beta1.DnsZone
	if kzone, ok = obj.(*v1beta1.DnsZone); !ok {
		return false, errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var rzone *zoneAndRecords
	if rzone, ok = resource.(*zoneAndRecords); !ok {
		return false, errors.Errorf("received '%s' resource instead of '*zoneAndRecords'", reflect.TypeOf(resource))
	}

	// Compare
	if rzone.Zone.Id != kzone.Status.Id {
		return true, nil
	}

	return false, nil
}

func (a *ResourceAdapter) UpdateObjectStatus(obj interface{}, resource interface{}) error {
	var ok bool

	var kzone *v1beta1.DnsZone
	if kzone, ok = obj.(*v1beta1.DnsZone); !ok {
		return errors.Errorf("received '%s' object instead of '*v1beta1.DnsZone'", reflect.TypeOf(obj))
	}

	var rzone *zoneAndRecords
	if rzone, ok = resource.(*zoneAndRecords); !ok {
		return errors.Errorf("received '%s' resource instead of '*zoneAndRecords'", reflect.TypeOf(resource))
	}

	kzone.Status.Id = rzone.Zone.Id
	return nil
}
