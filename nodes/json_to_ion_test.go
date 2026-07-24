package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// Composes JsonToIon with IonToJson (the node under test in
// ion_to_json_test.go) as a genuine cross-node consistency oracle: a JSON
// document containing only the types JSON and Ion share (null/bool/number/
// string/array/object) must survive JsonToIon -> IonToJson unchanged. This
// is stronger than a same-function round-trip because it exercises both
// nodes' independently-written implementations against each other.
func TestJsonToIon_RoundTripsThroughIonToJson(t *testing.T) {
	ctx := newTestContext(t)
	const original = `{"a":1,"b":"x","c":[true,false,null],"d":2.5}`

	toIon, err := nodes.JsonToIon(context.Background(), ctx, &gen.JsonDoc{Json: original})
	if err != nil {
		t.Fatalf("unexpected error from JsonToIon: %v", err)
	}
	if toIon.Error != "" {
		t.Fatalf("unexpected node error from JsonToIon: %s", toIon.Error)
	}

	backToJSON, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(toIon.Text)})
	if err != nil {
		t.Fatalf("unexpected error from IonToJson: %v", err)
	}
	if backToJSON.Error != "" {
		t.Fatalf("unexpected node error from IonToJson: %s", backToJSON.Error)
	}

	if backToJSON.Json != original {
		t.Fatalf("round trip mismatch:\n original: %s\n produced ion: %q\n back to json: %s", original, toIon.Text, backToJSON.Json)
	}
}

// Field order in the source JSON is preserved in the produced Ion struct —
// spot-checked directly on JsonToIon's own output (not via a round trip).
func TestJsonToIon_PreservesFieldOrder(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.JsonToIon(context.Background(), ctx, &gen.JsonDoc{Json: `{"z":1,"a":2,"m":3}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "{z:1,a:2,m:3}\n"
	if got.Text != want {
		t.Fatalf("text mismatch:\n got:  %q\n want: %q", got.Text, want)
	}
}

// A JSON integer too large for int64 must still convert (Ion ints are
// arbitrary precision) instead of overflowing or erroring.
func TestJsonToIon_BigIntegerPreservesPrecision(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.JsonToIon(context.Background(), ctx, &gen.JsonDoc{Json: `123456789012345678901234567890`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}
	want := "123456789012345678901234567890\n"
	if got.Text != want {
		t.Fatalf("text mismatch:\n got:  %q\n want: %q", got.Text, want)
	}
}

func TestJsonToIon_MissingJsonIsError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.JsonToIon(context.Background(), ctx, &gen.JsonDoc{})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for empty json, got none: %+v", got)
	}
}

func TestJsonToIon_MalformedJsonIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.JsonToIon(context.Background(), ctx, &gen.JsonDoc{Json: `{"a":1,`})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed json, got none: %+v", got)
	}
	if got.Text != "" {
		t.Fatalf("expected empty Text alongside a set Error, got %q", got.Text)
	}
}
