package nodes_test

import (
	"context"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

// Common types (int, string, list, bool, null, float) map onto JSON exactly
// as any JSON encoder would produce them — this is a hand-computed known
// value any independent JSON library agrees with. Note "2.5e0" (with an
// explicit exponent) is required for this to be an Ion FLOAT — Ion's bare
// "2.5" is a DECIMAL (see TestIonToJson_DecimalBecomesAJSONString below),
// one of the sharper differences between Ion's and JSON's number handling.
func TestIonToJson_CommonTypes(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:"x",c:[true,false,null],d:2.5e0}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}

	want := `{"a":1,"b":"x","c":[true,false,null],"d":2.5}`
	if got.Json != want {
		t.Fatalf("json mismatch:\n got:  %s\n want: %s", got.Json, want)
	}
}

// Ion decimal has exact, arbitrary-precision, trailing-zero-significant
// semantics that a JSON number cannot represent without silently changing
// the value (and Ion's 'd' exponent marker isn't even valid JSON number
// syntax) — so decimals become JSON STRINGS holding their exact Ion text,
// never JSON numbers. This test proves that documented, deliberate lossiness
// actually happens, rather than just asserting it in prose.
func TestIonToJson_DecimalBecomesAJSONString(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`19.990`)}) // trailing zero is significant in Ion decimal
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"19.990"`
	if got.Json != want {
		t.Fatalf("decimal conversion mismatch:\n got:  %s\n want: %s (a JSON string, not a number)", got.Json, want)
	}
}

// Ion timestamp has no JSON equivalent type, so it becomes a JSON string
// holding the exact ISO-8601 text.
func TestIonToJson_TimestampBecomesAJSONString(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`2026-07-24T12:00:00Z`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"2026-07-24T12:00:00Z"`
	if got.Json != want {
		t.Fatalf("timestamp conversion mismatch:\n got:  %s\n want: %s", got.Json, want)
	}
}

// Annotations are dropped — JSON has no equivalent concept. This proves the
// drop actually happens (not just documented): the annotated value
// `meters::5` converts to the bare JSON number 5, with "meters" nowhere in
// the output.
func TestIonToJson_AnnotationsAreDropped(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`meters::5`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `5`
	if got.Json != want {
		t.Fatalf("annotation-drop mismatch:\n got:  %s\n want: %s (annotation must not appear)", got.Json, want)
	}
}

// Blob content becomes a base64-encoded JSON string.
func TestIonToJson_BlobBecomesBase64String(t *testing.T) {
	ctx := newTestContext(t)

	// {{aGVsbG8=}} is an Ion blob literal containing base64 for "hello".
	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`{{aGVsbG8=}}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `"aGVsbG8="`
	if got.Json != want {
		t.Fatalf("blob conversion mismatch:\n got:  %s\n want: %s", got.Json, want)
	}
}

func TestIonToJson_MalformedInputIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.IonToJson(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,`)})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion, got none: %+v", got)
	}
	if got.Json != "" {
		t.Fatalf("expected empty Json alongside a set Error, got %q", got.Json)
	}
}
