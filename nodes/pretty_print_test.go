package nodes_test

import (
	"context"
	"strings"
	"testing"

	gen "christiangeorgelucas/ion-tools/gen"
	"christiangeorgelucas/ion-tools/nodes"
)

func TestPrettyPrint_IndentsNestedContainers(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.PrettyPrint(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,3]}`)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error: %s", got.Error)
	}
	if !strings.Contains(got.Text, "\n") {
		t.Fatalf("expected pretty-printed output to contain newlines, got %q", got.Text)
	}
	if got.Text == `{a:1,b:[1,2,3]}` {
		t.Fatalf("pretty-printed output should not equal the compact single-line form")
	}

	// The pretty-printed text must still be parseable Ion carrying the same
	// data — cross-checked by re-running it through InspectStructure.
	report, err := nodes.InspectStructure(context.Background(), ctx, &gen.IonInput{Data: []byte(got.Text)})
	if err != nil {
		t.Fatalf("unexpected error re-parsing pretty-printed output: %v", err)
	}
	if report.Error != "" || report.ValueCount != 6 {
		t.Fatalf("pretty-printed output did not re-parse to the same shape (want 6 values: struct, a, list, 1, 2, 3): %+v", report)
	}
}

func TestPrettyPrint_AcceptsBinaryInputToo(t *testing.T) {
	ctx := newTestContext(t)

	bin, err := nodes.TextToBinary(context.Background(), ctx, &gen.IonText{Text: `{a:1}`})
	if err != nil || bin.Error != "" {
		t.Fatalf("setup failed: err=%v node_err=%s", err, bin.Error)
	}

	got, err := nodes.PrettyPrint(context.Background(), ctx, &gen.IonInput{Data: bin.Data})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Error != "" {
		t.Fatalf("unexpected node error pretty-printing binary input: %s", got.Error)
	}
	if !strings.Contains(got.Text, "a: 1") {
		t.Fatalf("expected pretty-printed text to contain the field, got %q", got.Text)
	}
}

func TestPrettyPrint_MalformedInputIsStructuredError(t *testing.T) {
	ctx := newTestContext(t)

	got, err := nodes.PrettyPrint(context.Background(), ctx, &gen.IonInput{Data: []byte(`{a:1,b:[1,2,`)})
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if got.Error == "" {
		t.Fatalf("expected a structured error for malformed Ion, got none: %+v", got)
	}
}
