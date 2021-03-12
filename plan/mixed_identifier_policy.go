package plan

import (
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/external-dns/endpoint"
)

var (
	conflictSkipTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "external_dns",
			Subsystem: "policy",
			Name:      "conflict_skip_total",
			Help:      "Number of skipped records.",
		},
	)
)

func init() {
	prometheus.MustRegister(conflictSkipTotal)
}

// NoMixedIdentifierPolicy allows to avoid creating mixed record types (with/without identifier)
type NoMixedIdentifierPolicy struct{}

// Apply looks at the requested changes and drops all record creation that would lead to mixed types (with/without identifier)
func (p *NoMixedIdentifierPolicy) Apply(changes *Changes) *Changes {
	knownWeighted := map[string]*endpoint.Endpoint{}
	knownUnweighted := map[string]*endpoint.Endpoint{}
	sameOwner := func(ep1, ep2 *endpoint.Endpoint) bool {
		if owner1, ok := ep1.Labels[endpoint.OwnerLabelKey]; ok {
			if owner2, ok := ep2.Labels[endpoint.OwnerLabelKey]; ok {
				return owner1 == owner2
			}
			return false
		}
		return true
	}
	for _, ep := range changes.Current {
		if ep.SetIdentifier == "" {
			knownUnweighted[ep.DNSName] = ep
		} else {
			knownWeighted[ep.DNSName] = ep
		}

	}
	for _, candidate := range changes.Delete {
		if candidate.SetIdentifier == "" {
			if ep, ok := knownUnweighted[candidate.DNSName]; ok && sameOwner(candidate, ep) {
				delete(knownUnweighted, candidate.DNSName)
			}
		} else {
			if ep, ok := knownWeighted[candidate.DNSName]; ok && sameOwner(candidate, ep) {
				delete(knownWeighted, candidate.DNSName)
			}
		}

	}
	create := []*endpoint.Endpoint{}
	for _, ep := range changes.Create {
		if ep.SetIdentifier == "" {
			if known, ok := knownWeighted[ep.DNSName]; !ok {
				create = append(create, ep)
			} else {
				conflictSkipTotal.Inc()
				log.Infof(`Skipping endpoint creation %v without identifier because there exists a conflicting record without identifier: %v`, ep, known)
			}
		} else {
			if known, ok := knownUnweighted[ep.DNSName]; !ok {
				create = append(create, ep)
			} else {
				conflictSkipTotal.Inc()
				log.Infof(`Skipping endpoint creation %v with identifier because there exists a conflicting record with identifier: %v`, ep, known)
			}
		}
	}
	return &Changes{
		Create:    create,
		UpdateOld: changes.UpdateOld,
		UpdateNew: changes.UpdateNew,
		Delete:    changes.Delete,
		Current:   changes.Current,
	}
}
