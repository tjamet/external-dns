package plan

import (
	"testing"

	"sigs.k8s.io/external-dns/endpoint"
)

func TestNoMixedPolicy(t *testing.T) {
	// empty list of records
	empty := []*endpoint.Endpoint{}
	// a simple entry
	fooV1 := []*endpoint.Endpoint{{DNSName: "foo", Targets: endpoint.Targets{"v1"}}}
	// the same entry but with different target
	fooV2 := []*endpoint.Endpoint{{DNSName: "foo", Targets: endpoint.Targets{"v2"}}}
	// another two simple entries
	bar := []*endpoint.Endpoint{{DNSName: "bar", Targets: endpoint.Targets{"v1"}}}
	baz := []*endpoint.Endpoint{{DNSName: "baz", Targets: endpoint.Targets{"v1"}}}

	identified := []*endpoint.Endpoint{{DNSName: "multi-cluster", SetIdentifier: "identifier", Targets: endpoint.Targets{"identified"}}}
	nonIdentified := []*endpoint.Endpoint{{DNSName: "multi-cluster", Targets: endpoint.Targets{"non-identified"}}}

	for _, tc := range []struct {
		policy   Policy
		changes  *Changes
		expected *Changes
	}{
		{
			// NoMixedIdentifierPolicy avoids creating conflicting records with and without identifiers
			&NoMixedIdentifierPolicy{},
			&Changes{Create: baz, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar},
			&Changes{Create: baz, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar},
		},
		{
			// NoMixedIdentifierPolicy avoids creating conflicting records with and without identifiers
			&NoMixedIdentifierPolicy{},
			&Changes{Create: identified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar, Current: nonIdentified},
			&Changes{Create: empty, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar},
		},
		{
			// NoMixedIdentifierPolicy avoids creating conflicting records with and without identifiers
			&NoMixedIdentifierPolicy{},
			&Changes{Create: identified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: nonIdentified, Current: nonIdentified},
			&Changes{Create: identified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: nonIdentified},
		},
		{
			// NoMixedIdentifierPolicy avoids creating conflicting records with and without identifiers
			&NoMixedIdentifierPolicy{},
			&Changes{Create: nonIdentified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar, Current: identified},
			&Changes{Create: empty, UpdateOld: fooV1, UpdateNew: fooV2, Delete: bar},
		},
		{
			// NoMixedIdentifierPolicy avoids creating conflicting records with and without identifiers
			&NoMixedIdentifierPolicy{},
			&Changes{Create: nonIdentified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: identified, Current: identified},
			&Changes{Create: nonIdentified, UpdateOld: fooV1, UpdateNew: fooV2, Delete: identified},
		},
	} {
		// apply policy
		changes := tc.policy.Apply(tc.changes)

		// validate changes after applying policy
		validateEntries(t, changes.Create, tc.expected.Create)
		validateEntries(t, changes.UpdateOld, tc.expected.UpdateOld)
		validateEntries(t, changes.UpdateNew, tc.expected.UpdateNew)
		validateEntries(t, changes.Delete, tc.expected.Delete)
	}
}
