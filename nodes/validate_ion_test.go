package nodes_test

import (
	"context"
	_ "embed"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// testdata/ion_tests_strings.ion is copied verbatim from
// amazon-ion/ion-tests (Apache-2.0; see testdata/ION-TESTS-LICENSE), the
// Ion team's own official cross-implementation conformance corpus —
// iontestdata/good/strings.ion, a file the spec guarantees is well-formed
// Ion. Its 20 top-level values were independently cross-checked against
// Python's amazon.ion (`len(list(amazon.ion.simpleion.loads(data,
// single_value=False)))` also returns 20) before being hard-coded below, so
// this is an oracle test against BOTH the official conformance suite and a
// second independent implementation, not just this package's own opinion of
// what "valid" means.
//
//go:embed testdata/ion_tests_strings.ion
var officialGoodStringsFixture []byte

func TestValidateIon_OfficialConformanceCorpusIsValid(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ValidateIon(context.Background(), ctx, &gen.IonInput{Data: officialGoodStringsFixture})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid {
		t.Fatalf("expected the official ion-tests good/strings.ion fixture to validate, got invalid: %s", got.Error)
	}
	if got.ValueCount != 20 {
		t.Fatalf("value_count = %d, want 20 (independently confirmed via Python amazon.ion)", got.ValueCount)
	}
}

func TestValidateIon_WellFormedInput(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ValidateIon(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1} [1,2,3] "hello"`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Valid {
		t.Fatalf("expected valid, got invalid: %s", got.Error)
	}
	if got.ValueCount != 3 {
		t.Fatalf("value_count = %d, want 3", got.ValueCount)
	}
}

func TestValidateIon_MalformedInputIsInvalid(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ValidateIon(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,`)})
	if err != nil {
		t.Fatalf("unexpected Go error (should be a structured field, not a thrown error): %v", err)
	}
	if got.Valid {
		t.Fatalf("expected invalid for unterminated Ion, got valid")
	}
	if got.Error == "" {
		t.Fatalf("expected a non-empty error message alongside valid=false")
	}
}

func TestValidateIon_MissingDataIsInvalid(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.ValidateIon(context.Background(), ctx, &gen.IonInput{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Valid {
		t.Fatalf("expected invalid for empty input, got valid")
	}
}
