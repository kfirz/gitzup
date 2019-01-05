package gcp

import (
	"context"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/dns/v1"
	"strings"
)

func CreateComputeClient() (*compute.Service, error) {
	// TODO: think about which scopes we're using to authenticate to GCP
	googleClient, err := google.DefaultClient(context.TODO(), compute.CloudPlatformScope)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP compute client")
	}

	computeClient, err := compute.New(googleClient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP compute client")
	}

	return computeClient, nil
}

func CreateDnsClient() (*dns.Service, error) {
	// TODO: think about which scopes we're using to authenticate to GCP
	googleClient, err := google.DefaultClient(context.TODO(), compute.CloudPlatformScope)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP DNS client")
	}

	dnsClient, err := dns.New(googleClient)
	if err != nil {
		return nil, errors.Wrapf(err, "failed creating GCP DNS client")
	}

	return dnsClient, nil
}

func WaitForComputeOperation(projectId string, op *compute.Operation) error {
	svc, err := CreateComputeClient()
	if err != nil {
		return errors.Wrapf(err, "failed creating GCP compute client")
	}

	if op.Zone != "" {
		lastSlashIndex := strings.LastIndex(op.Zone, "/")
		zone := op.Zone[lastSlashIndex + 1:]
		for {
			resp, err := svc.ZoneOperations.Get(projectId, zone, op.Name).Do()
			if err != nil {
				return errors.Wrapf(err, "failed getting zonal operation status")
			} else if strings.ToLower(resp.Status) == "done" {
				if resp.Error != nil {
					return errors.Errorf("zonal operation failed: %+v", resp.Error)
				}
				return nil
			}
		}
	} else if op.Region != "" {
		lastSlashIndex := strings.LastIndex(op.Region, "/")
		region := op.Region[lastSlashIndex + 1:]
		for {
			resp, err := svc.RegionOperations.Get(projectId, region, op.Name).Do()
			if err != nil {
				return errors.Wrapf(err, "failed getting regional operation status")
			} else if strings.ToLower(resp.Status) == "done" {
				if resp.Error != nil {
					return errors.Errorf("regional operation failed: %+v", resp.Error)
				}
				return nil
			}
		}
	} else {
		for {
			resp, err := svc.GlobalOperations.Get(projectId, op.Name).Do()
			if err != nil {
				return errors.Wrapf(err, "failed getting global operation status")
			} else if strings.ToLower(resp.Status) == "done" {
				if resp.Error != nil {
					return errors.Errorf("global operation failed: %+v", resp.Error)
				}
				return nil
			}
		}
	}
}

func WaitForDnsOperation(projectId string, op *dns.Operation) error {
	svc, err := CreateDnsClient()
	if err != nil {
		return errors.Wrapf(err, "failed creating GCP DNS client")
	}

	for {
		resp, err := svc.ManagedZoneOperations.Get(projectId, op.ZoneContext.OldValue.Name, op.Id).Do()
		if err != nil {
			return errors.Wrapf(err, "failed getting DNS operation status")
		} else if strings.ToLower(resp.Status) == "done" {
			return nil
		}
	}
}
