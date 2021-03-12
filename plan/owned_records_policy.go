package plan

import (
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

type OwnedRecordsPolicy struct {
	OwnerID string
}

func (pol *OwnedRecordsPolicy) Apply(changes *Changes) *Changes {
	return &Changes{
		Create:    changes.Create,
		UpdateNew: filterOwnedRecords(pol.OwnerID, changes.UpdateNew),
		UpdateOld: filterOwnedRecords(pol.OwnerID, changes.UpdateOld),
		Delete:    filterOwnedRecords(pol.OwnerID, changes.Delete),
	}
}

func filterOwnedRecords(ownerID string, eps []*endpoint.Endpoint) []*endpoint.Endpoint {
	filtered := []*endpoint.Endpoint{}
	for _, ep := range eps {
		if endpointOwner, ok := ep.Labels[endpoint.OwnerLabelKey]; !ok || endpointOwner != ownerID {
			log.Debugf(`Skipping endpoint %v because owner id does not match, found: "%s", required: "%s"`, ep, endpointOwner, ownerID)
			continue
		}
		filtered = append(filtered, ep)
	}
	return filtered
}
